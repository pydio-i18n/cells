package modifiers

import (
	"fmt"
	"github.com/pydio/cells/v4/common/client/grpc"
	"net/http"
	"strconv"

	"github.com/emicklei/go-restful"
	"github.com/gorilla/sessions"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/pydio/cells/v4/common"
	"github.com/pydio/cells/v4/common/broker"
	"github.com/pydio/cells/v4/common/log"
	"github.com/pydio/cells/v4/common/proto/idm"
	"github.com/pydio/cells/v4/common/proto/rest"
	"github.com/pydio/cells/v4/common/proto/service"
	"github.com/pydio/cells/v4/common/service/errors"
	"github.com/pydio/cells/v4/common/service/frontend"
	"github.com/pydio/cells/v4/common/utils/i18n"
	"github.com/pydio/cells/v4/common/utils/permissions"
	json "github.com/pydio/cells/v4/x/jsonx"
)

// LoginSuccessWrapper wraps functionalities after user was successfully logged in
func LoginSuccessWrapper(middleware frontend.AuthMiddleware) frontend.AuthMiddleware {
	return func(req *restful.Request, rsp *restful.Response, in *rest.FrontSessionRequest, out *rest.FrontSessionResponse, session *sessions.Session) error {
		if a, ok := in.AuthInfo["type"]; !ok || a != "credentials" && a != "authorization_code" { // Ignore this middleware
			return middleware(req, rsp, in, out, session)
		}
		// BEFORE MIDDLEWARE

		// MIDDLEWARE
		err := middleware(req, rsp, in, out, session)
		if err != nil {
			return err
		}

		// AFTER MIDDLEWARE

		// retrieving user
		username, ok := in.AuthInfo["login"]
		if !ok {
			return errors.New("user.not_found", "User id not found", http.StatusNotFound)
		}

		ctx := req.Request.Context()

		// Searching user for attributes
		user, err := permissions.SearchUniqueUser(ctx, username, "")
		if err != nil {
			return err
		}

		// If user is "hidden", he is not allowed to log on the main interface
		if user.Attributes != nil {
			_, mini := session.Values["minisite"]
			if user.IsHidden() && !mini {
				log.Auditer(ctx).Error(
					"Hidden user ["+user.Login+"] tried to log in on main interface",
					log.GetAuditId(common.AuditLoginPolicyDenial),
					zap.String(common.KeyUserUuid, user.Uuid),
				)
				log.Logger(ctx).Error("Denied login for hidden user " + user.Login + " on main interface")
				return errors.Unauthorized("hidden.user.nominisite", "You are not allowed to log in to this interface")
			}
		}

		// Checking user is locked
		if permissions.IsUserLocked(user) {
			log.Auditer(ctx).Error(
				"Locked user ["+user.Login+"] tried to log in.",
				log.GetAuditId(common.AuditLoginPolicyDenial),
				zap.String(common.KeyUserUuid, user.Uuid),
			)
			log.Logger(ctx).Error("lock denies login for "+user.Login, zap.Error(fmt.Errorf("blocked login")))
			return errors.Unauthorized(common.ServiceUser, "User "+user.Login+" has been blocked. Contact your sysadmin.")
		}

		// Reset failed connections
		if user.Attributes != nil {
			if _, ok := user.Attributes["failedConnections"]; ok {
				log.Logger(ctx).Info("[WrapWithUserLocks] Resetting user failedConnections", user.ZapLogin())
				userClient := idm.NewUserServiceClient(grpc.NewClientConn(common.ServiceUser))
				delete(user.Attributes, "failedConnections")
				userClient.CreateUser(ctx, &idm.CreateUserRequest{User: user})
			}
		}

		// Checking policies
		cli := idm.NewPolicyEngineServiceClient(grpc.NewClientConn(common.ServicePolicy))
		policyContext := make(map[string]string)
		permissions.PolicyContextFromMetadata(policyContext, ctx)
		subjects := permissions.PolicyRequestSubjectsFromUser(user)

		// Check all subjects, if one has deny return false
		policyRequest := &idm.PolicyEngineRequest{
			Subjects: subjects,
			Resource: "oidc",
			Action:   "login",
			Context:  policyContext,
		}

		if resp, err := cli.IsAllowed(ctx, policyRequest); err != nil || !resp.Allowed {
			log.Auditer(ctx).Error(
				"Policy denied login to ["+user.Login+"]",
				log.GetAuditId(common.AuditLoginPolicyDenial),
				zap.String(common.KeyUserUuid, user.Uuid),
				zap.Any(common.KeyPolicyRequest, policyRequest),
				zap.Error(err),
			)
			log.Logger(ctx).Error("policy denies login for request", zap.Any(common.KeyPolicyRequest, policyRequest), zap.Error(err))
			return errors.Unauthorized(common.ServiceUser, "User "+user.Login+" is not authorized to log in")
		}

		if lang, ok := in.AuthInfo["lang"]; ok {
			if _, o := i18n.AvailableLanguages[lang]; o {
				aclClient := idm.NewACLServiceClient(grpc.NewClientConn(common.ServiceAcl))
				// Remove previous value if any
				delQ, _ := anypb.New(&idm.ACLSingleQuery{RoleIDs: []string{user.GetUuid()}, Actions: []*idm.ACLAction{{Name: "parameter:core.conf:lang"}}, WorkspaceIDs: []string{"PYDIO_REPO_SCOPE_ALL"}})
				aclClient.DeleteACL(ctx, &idm.DeleteACLRequest{Query: &service.Query{SubQueries: []*anypb.Any{delQ}}} /* TODO V4, client.WithRequestTimeout(500*time.Millisecond)*/)
				// Insert new ACL with language value
				_, e := aclClient.CreateACL(ctx, &idm.CreateACLRequest{ACL: &idm.ACL{
					Action:      &idm.ACLAction{Name: "parameter:core.conf:lang", Value: lang},
					RoleID:      user.GetUuid(),
					WorkspaceID: "PYDIO_REPO_SCOPE_ALL",
				}} /* TODO V4, client.WithRequestTimeout(500*time.Millisecond)*/)
				if e != nil {
					log.Logger(ctx).Error("Cannot update language for user", user.ZapLogin(), zap.String("lang", lang), zap.Error(e))
				} else {
					log.Logger(ctx).Info("Updated language for "+user.GetLogin()+" to "+lang, user.ZapLogin(), zap.String("lang", lang))
				}
			}
		}

		broker.MustPublish(ctx, common.TopicIdmEvent, &idm.ChangeEvent{
			Type: idm.ChangeEventType_LOGIN,
			User: user,
		})

		return nil
	}
}

// LoginFailedWrapper wraps functionalities after user failed to log in
func LoginFailedWrapper(middleware frontend.AuthMiddleware) frontend.AuthMiddleware {
	return func(req *restful.Request, rsp *restful.Response, in *rest.FrontSessionRequest, out *rest.FrontSessionResponse, session *sessions.Session) error {
		if a, ok := in.AuthInfo["type"]; !ok || a != "credentials" && a != "external" { // Ignore this middleware
			return middleware(req, rsp, in, out, session)
		}

		// MIDDLEWARE
		err := middleware(req, rsp, in, out, session)
		if err == nil {
			return nil
		}

		//fmt.Println("Login failed with ", err)

		ctx := req.Request.Context()

		username := in.AuthInfo["login"]

		const maxFailedLogins = 10

		// Searching user for attributes
		user, _ := permissions.SearchUniqueUser(ctx, username, "")

		if user == nil {
			return errors.New("login.failed", "Login failed", http.StatusUnauthorized)
		}

		// double check if user was already locked to reduce work load
		if permissions.IsUserLocked(user) {
			msg := fmt.Sprintf("locked user %s is still trying to connect", user.GetLogin())
			log.Logger(ctx).Warn(msg, user.ZapLogin())
			return errors.New("user.locked", "User is locked - Please contact your admin", http.StatusUnauthorized)
		}

		var failedInt int64
		if user.Attributes == nil {
			user.Attributes = make(map[string]string)
		}
		if failed, ok := user.Attributes["failedConnections"]; ok {
			failedInt, _ = strconv.ParseInt(failed, 10, 32)
		}
		failedInt++
		user.Attributes["failedConnections"] = fmt.Sprintf("%d", failedInt)
		hardLock := false

		if failedInt >= maxFailedLogins {
			// Set lock via attributes
			var locks []string
			if l, ok := user.Attributes["locks"]; ok {
				var existingLocks []string
				if e := json.Unmarshal([]byte(l), &existingLocks); e == nil {
					for _, lock := range existingLocks {
						if lock != "logout" {
							locks = append(locks, lock)
						}
					}
				}
			}
			locks = append(locks, "logout")
			data, _ := json.Marshal(locks)
			user.Attributes["locks"] = string(data)
			msg := fmt.Sprintf("Locked user [%s] after %d failed connections", user.GetLogin(), maxFailedLogins)
			log.Logger(ctx).Error(msg, user.ZapLogin())
			log.Auditer(ctx).Error(
				msg,
				log.GetAuditId(common.AuditLockUser),
				user.ZapLogin(),
				zap.String(common.KeyUserUuid, user.GetUuid()),
			)
			hardLock = true
		}

		log.Logger(ctx).Debug(fmt.Sprintf("[WrapWithUserLocks] Updating failed connection number for user [%s]", user.GetLogin()), user.ZapLogin())
		userClient := idm.NewUserServiceClient(grpc.NewClientConn(common.ServiceUser))
		if _, e := userClient.CreateUser(ctx, &idm.CreateUserRequest{User: user}); e != nil {
			log.Logger(ctx).Error("could not store failedConnection for user", zap.Error(e))
		}

		// Replacing error not to give any hint
		if hardLock {
			return errors.New("user.locked", "User is locked - Please contact your admin", http.StatusUnauthorized)
		} else {
			return errors.New("login.failed", "Login failed", http.StatusUnauthorized)
		}
	}
}
