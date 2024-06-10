/*
 * Copyright (c) 2019-2021. Abstrium SAS <team (at) pydio.com>
 * This file is part of Pydio Cells.
 *
 * Pydio Cells is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * Pydio Cells is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with Pydio Cells.  If not, see <http://www.gnu.org/licenses/>.
 *
 * The latest code can be found at <https://pydio.com>.
 */

package auth

import (
	"context"
	"net/url"
	"sort"
	"strings"

	"go.uber.org/zap"
	"golang.org/x/oauth2"

	"github.com/pydio/cells/v4/common"
	"github.com/pydio/cells/v4/common/auth/claim"
	errors3 "github.com/pydio/cells/v4/common/errors"
	"github.com/pydio/cells/v4/common/log"
	"github.com/pydio/cells/v4/common/proto/idm"
	"github.com/pydio/cells/v4/common/utils/permissions"
	"github.com/pydio/cells/v4/common/utils/propagator"
)

type ProviderType int

const (
	ProviderTypeOry ProviderType = iota
	ProviderTypeGrpc
	ProviderTypePAT
)

type Provider interface {
	GetType() ProviderType
}

type Verifier interface {
	Verify(context.Context, string) (IDToken, error)
}

type ContextVerifier interface {
	Verify(ctx context.Context, user *idm.User) error
}

type Exchanger interface {
	Exchange(context.Context, string, string) (*oauth2.Token, error)
}

// TokenOption is an AuthCodeOption is passed to Config.AuthCodeURL.
type TokenOption interface {
	setValue(url.Values)
}

type setParam struct{ k, v string }

func (p setParam) setValue(m url.Values) { m.Set(p.k, p.v) }

// SetChallenge builds a TokenOption which passes key/value parameters
// to a provider's token exchange endpoint.
func SetChallenge(value string) TokenOption {
	return setParam{"challenge", value}
}

// SetAccessToken builds a TokenOption for passing the access token.
func SetAccessToken(value string) TokenOption {
	return setParam{"access_token", value}
}

// SetRefreshToken builds a TokenOption for passing the refresh_token.
func SetRefreshToken(value string) TokenOption {
	return setParam{"refresh_token", value}
}

type PasswordCredentialsTokenExchanger interface {
	PasswordCredentialsToken(context.Context, string, string) (*oauth2.Token, error)
}

type PasswordCredentialsCodeExchanger interface {
	PasswordCredentialsCode(context.Context, string, string, ...TokenOption) (string, error)
}

type LoginChallengeCodeExchanger interface {
	LoginChallengeCode(context.Context, claim.Claims, ...TokenOption) (string, error)
}

type LogoutProvider interface {
	Logout(context.Context, string, string, string, ...TokenOption) error
}

type IDToken interface {
	Claims(interface{}) error
	ScopedClaims(claims *claim.Claims) error
}

var (
	providers        []Provider
	contextVerifiers []ContextVerifier
)

// AddContextVerifier registers an additional verifier
func AddContextVerifier(v ContextVerifier) {
	contextVerifiers = append(contextVerifiers, v)
}

// VerifyContext ranges over registered ContextVerifiers and check if one of them returns an error.
func VerifyContext(ctx context.Context, user *idm.User) error {
	for _, v := range contextVerifiers {
		if err := v.Verify(ctx, user); err != nil {
			return err
		}
	}
	return nil
}

type JWTVerifier struct {
	types []ProviderType
}

// DefaultJWTVerifier creates a ready to use JWTVerifier
func DefaultJWTVerifier() *JWTVerifier {
	return &JWTVerifier{
		types: []ProviderType{ProviderTypeGrpc, ProviderTypePAT},
	}
}

func (j *JWTVerifier) getProviders() []Provider {
	var res []Provider
	for _, provider := range providers {
		for _, t := range j.types {
			if provider.GetType() == t {
				res = append(res, provider)
			}
		}
	}

	return res
}

func (j *JWTVerifier) loadClaims(ctx context.Context, token IDToken, claims *claim.Claims) error {

	// Extract custom claims
	if err := token.Claims(claims); err != nil {
		log.Logger(ctx).Error("cannot extract custom claims from idToken", zap.Error(err))
		return err
	}

	if err := token.ScopedClaims(claims); err != nil {
		log.Logger(ctx).Error("cannot extract custom Scopes from claims", zap.Error(err))
	}

	// Search by name or by email
	var user *idm.User

	// Search by subject
	if claims.Subject != "" {
		if u, err := permissions.SearchUniqueUser(ctx, "", claims.Subject); err == nil {
			user = u
		} else if !errors3.Is(err, errors3.UserNotFound) {
			return err
		}
	} else if claims.Name != "" {
		if u, err := permissions.SearchUniqueUser(ctx, claims.Name, ""); err == nil {
			user = u
		} else if !errors3.Is(err, errors3.UserNotFound) {
			return err
		}
	}

	if user == nil && claims.Email != "" {
		if u, err := permissions.SearchUniqueUser(ctx, claims.Email, ""); err == nil {
			user = u
			// Now replace claims.Name
			claims.Name = claims.Email
		} else {
			return err
		}
	}
	if user == nil {
		return errors3.WithMessage(errors3.UserNotFound, "user not found neither by name or email")
	}

	// Check underlying verifiers
	if e := VerifyContext(ctx, user); e != nil {
		return errors3.Tag(e, errors3.StatusUnauthorized) // errors2.Unauthorized("user.context", e.Error())
	}

	displayName, ok := user.Attributes["displayName"]
	if !ok {
		displayName = ""
	}

	profile, ok := user.Attributes["profile"]
	if !ok {
		profile = "standard"
	}

	var roles []string
	for _, role := range user.Roles {
		roles = append(roles, role.Uuid)
	}

	claims.Name = user.Login
	claims.DisplayName = displayName
	claims.Profile = profile
	claims.Roles = strings.Join(roles, ",")
	claims.GroupPath = user.GroupPath

	return nil
}

func (j *JWTVerifier) verifyTokenWithRetry(ctx context.Context, rawIDToken string, isRetry bool) (IDToken, error) {

	var idToken IDToken
	var err error

	for _, provider := range j.getProviders() {
		verifier, ok := provider.(Verifier)
		if !ok {
			continue
		}

		idToken, err = verifier.Verify(ctx, rawIDToken)
		if err == nil {
			break
		}

		log.Logger(ctx).Debug("jwt rawIdToken verify: failed, trying next", zap.Error(err))
	}

	if (idToken == nil || err != nil) && !isRetry {
		return j.verifyTokenWithRetry(ctx, rawIDToken, true)
	}

	if idToken == nil {
		return nil, errors3.WithStack(errors3.EmptyIDToken)
	}

	return idToken, nil
}

// Exchange retrieves an oauth2 Token from a code.
func (j *JWTVerifier) Exchange(ctx context.Context, code, codeVerifier string) (*oauth2.Token, error) {
	var oauth2Token *oauth2.Token
	var err error

	for _, provider := range j.getProviders() {
		exch, ok := provider.(Exchanger)
		if !ok {
			continue
		}

		// Verify state and errors.
		oauth2Token, err = exch.Exchange(ctx, code, codeVerifier)
		if err == nil {
			break
		}

	}

	if err != nil {
		return nil, err
	}

	return oauth2Token, nil
}

// Verify validates an existing JWT token against the OIDC service that issued it
func (j *JWTVerifier) Verify(ctx context.Context, rawIDToken string) (context.Context, claim.Claims, error) {

	idToken, err := j.verifyTokenWithRetry(ctx, rawIDToken, false)
	if err != nil {
		log.Logger(ctx).Debug("error verifying token", zap.String("token", rawIDToken), zap.Error(err))
		return ctx, claim.Claims{}, err
	}

	claims := &claim.Claims{}
	if err := j.loadClaims(ctx, idToken, claims); err != nil {
		log.Logger(ctx).Error("got a token but failed to load claims", zap.Error(err))
		return ctx, *claims, err
	}

	ctx = ContextFromClaims(ctx, *claims)

	return ctx, *claims, nil
}

// PasswordCredentialsToken will perform a call to the OIDC service with grantType "password"
// to get a valid token from a given user/pass credentials
func (j *JWTVerifier) PasswordCredentialsToken(ctx context.Context, userName string, password string) (*oauth2.Token, error) {

	var token *oauth2.Token
	var err error

	for _, provider := range j.getProviders() {
		recl, ok := provider.(PasswordCredentialsTokenExchanger)
		if !ok {
			continue
		}

		token, err = recl.PasswordCredentialsToken(ctx, userName, password)
		if err == nil {
			break
		}
	}
	if token == nil {
		if err == nil {
			err = errors3.WithStack(errors3.EmptyIDToken) // errors2.Unauthorized("empty.token", "could not validate password credentials")
		}
	}
	return token, err
}

// LoginChallengeCode will perform an implicit flow
// to get a valid code from given claims and challenge
func (j *JWTVerifier) LoginChallengeCode(ctx context.Context, claims claim.Claims, opts ...TokenOption) (string, error) {

	var code string
	var err error

	for _, provider := range j.getProviders() {
		p, ok := provider.(LoginChallengeCodeExchanger)
		if !ok {
			continue
		}

		code, err = p.LoginChallengeCode(ctx, claims, opts...)
		if err == nil {
			break
		}
	}

	return code, err
}

// PasswordCredentialsCode will perform an implicit flow
// to get a valid code from given claims and challenge
func (j *JWTVerifier) PasswordCredentialsCode(ctx context.Context, username, password string, opts ...TokenOption) (string, error) {

	var code string
	var err error

	for _, provider := range j.getProviders() {
		p, ok := provider.(PasswordCredentialsCodeExchanger)
		if !ok {
			continue
		}

		code, err = p.PasswordCredentialsCode(ctx, username, password, opts...)
		if err == nil {
			break
		}
	}

	return code, err
}

// Logout calls logout on underlying provider
func (j *JWTVerifier) Logout(ctx context.Context, url, subject, sessionID string, opts ...TokenOption) error {
	for _, provider := range j.getProviders() {
		p, ok := provider.(LogoutProvider)
		if !ok {
			continue
		}

		if err := p.Logout(ctx, url, subject, sessionID, opts...); err != nil {
			return err
		}
	}

	return nil
}

// WithImpersonate Add a fake Claims in context to impersonate a user.
func WithImpersonate(ctx context.Context, user *idm.User) context.Context {
	roles := make([]string, len(user.Roles))
	for _, r := range user.Roles {
		roles = append(roles, r.Uuid)
	}
	// Build Claims Now
	c := claim.Claims{
		Subject:   user.Uuid,
		Name:      user.Login,
		GroupPath: user.GroupPath,
		Roles:     strings.Join(roles, ","),
	}
	if user.Attributes != nil {
		if p, o := user.Attributes[idm.UserAttrProfile]; o {
			c.Profile = p
		}
		if e, o := user.Attributes[idm.UserAttrEmail]; o {
			c.Email = e
		}
		if a, o := user.Attributes[idm.UserAttrAuthSource]; o {
			c.AuthSource = a
		}
		if dn, o := user.Attributes[idm.UserAttrDisplayName]; o {
			c.DisplayName = dn
		}
	}
	ctx = propagator.WithAdditionalMetadata(ctx, map[string]string{common.PydioContextUserKey: user.Login})
	return context.WithValue(ctx, claim.ContextKey, c)
}

func addProvider(p Provider) {
	providers = append(providers, p)
	sortProviders()
}

func sortProviders() {
	sort.Slice(providers, func(i, j int) bool {
		return providers[i].GetType() < providers[j].GetType()
	})
}
