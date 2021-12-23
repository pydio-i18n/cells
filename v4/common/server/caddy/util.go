package caddy

import (
	"net/url"
	"path/filepath"

	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"

	"github.com/pydio/cells/v4/common"
	"github.com/pydio/cells/v4/common/caddy/maintenance"
	"github.com/pydio/cells/v4/common/config"
	"github.com/pydio/cells/v4/common/crypto/providers"
	"github.com/pydio/cells/v4/common/proto/install"
	"github.com/pydio/cells/v4/common/utils/statics"
)

var maintenanceDir string

// SitesToCaddyConfigs computes all SiteConf from all *install.ProxyConfig by analyzing
// TLSConfig, ReverseProxyURL and Maintenance fields values
func SitesToCaddyConfigs(sites []*install.ProxyConfig) (caddySites []SiteConf, er error) {
	for _, proxyConfig := range sites {
		if bc, er := computeSiteConf(proxyConfig); er == nil {
			caddySites = append(caddySites, bc)
		} else {
			return caddySites, er
		}
	}
	return caddySites, nil
}

// GetMaintenanceRoot provides a static root folder for serving maintenance page
func GetMaintenanceRoot() (string, error) {
	if maintenanceDir != "" {
		return maintenanceDir, nil
	}
	dir, err := statics.GetAssets("./maintenance/src")
	if err != nil {
		dir = filepath.Join(config.ApplicationWorkingDir(), "static", "maintenance")
		if _, _, err := statics.RestoreAssets(dir, maintenance.PydioMaintenanceBox, nil); err != nil {
			return "", errors.Wrap(err, "could not restore maintenance package")
		}
	}
	maintenanceDir = dir
	return dir, nil
}

func computeSiteConf(pc *install.ProxyConfig) (SiteConf, error) {
	bc := SiteConf{
		ProxyConfig: proto.Clone(pc).(*install.ProxyConfig),
	}
	if pc.ReverseProxyURL != "" {
		if u, e := url.Parse(pc.ReverseProxyURL); e == nil {
			bc.ExternalHost = u.Host
		}
	}
	if bc.TLSConfig == nil {
		for i, b := range bc.Binds {
			bc.Binds[i] = "http://" + b
		}
	} else {
		switch v := bc.TLSConfig.(type) {
		case *install.ProxyConfig_Certificate, *install.ProxyConfig_SelfSigned:
			certFile, keyFile, err := providers.LoadCertificates(pc)
			if err != nil {
				return bc, err
			}
			bc.TLSCert = certFile
			bc.TLSKey = keyFile
		case *install.ProxyConfig_LetsEncrypt:
			// todo v4 : is there something to do with AcceptEULA flag ? Was caddytls.Agreed = true
			caUrl := common.DefaultCaUrl
			if v.LetsEncrypt.StagingCA {
				caUrl = common.DefaultCaStagingUrl
			}
			bc.TLS = v.LetsEncrypt.Email + ` {
				ca ` + caUrl + `
			}`
		}
	}
	if bc.Maintenance {
		mDir, e := GetMaintenanceRoot()
		if e != nil {
			return bc, e
		}
		bc.WebRoot = mDir
	}
	return bc, nil
}
