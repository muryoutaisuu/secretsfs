package store

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/hashicorp/vault/api"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	sfsfh "github.com/muryoutaisuu/secretsfs/pkg/fusehelpers"
	vh "github.com/muryoutaisuu/vaulthelper"
	pfvault "github.com/postfinance/vault/kv"
)

// kv mount path
const KVMountPath = "secret/"

type VaultKv struct {
}

var _ = (Store)((*VaultKv)(nil))

func (s *VaultKv) GetSecret(spath string, ctx context.Context) (*Secret, error) {
	return getSecret(spath, ctx, true)
}

func getSecret(spath string, ctx context.Context, appendSubs bool) (*Secret, error) {
	u, err := sfsfh.GetUserFromContext(ctx)
	if err != nil {
		log.WithFields(log.Fields{
			"spath":      spath,
			"appendSubs": appendSubs,
			"error":      err}).Error("got error while getting user from context")
		return nil, err
	}
	log.WithFields(log.Fields{
		"spath":      spath,
		"appendSubs": appendSubs,
		"username":   u.Username}).Info("User accessing a secret")
	c, err := GetClient(ctx)
	if err != nil {
		log.WithFields(log.Fields{
			"spath":      spath,
			"pfclient":   c,
			"appendSubs": appendSubs,
			"error":      err}).Error("got error while getting vault client")
		return nil, err
	}
	t := vh.GetTypes(c, KVMountPath+spath)

	switch {
	case t[vh.CPath], t[vh.CSecret]:
		s := &Secret{
			Path: spath,
			Mode: sfsfh.DIRREAD,
		}

		// append keys as Subs, if type is CScret
		if t[vh.CSecret] && appendSubs {
			data, err := c.Read(KVMountPath + spath)
			if err != nil {
				log.WithFields(log.Fields{
					"spath":         spath,
					"pfclient":      c,
					"appendSubs":    appendSubs,
					"t":             t,
					"t[vh.CSecret]": t[vh.CSecret],
					"type":          "vh.CSecret",
					"storesecret":   s,
					"data":          data,
					"error":         err}).Warn("got error while getting vault secret with client and spath for adding as subs to store secret. Continuing...")
			} else {
				for k := range data {
					newsec := &Secret{
						Path: filepath.Join(spath, k),
						Mode: sfsfh.FILEREAD,
					}
					s.Subs = append(s.Subs, newsec)
				}
			}
		}

		// append Paths as Subs, if type is CPath
		if t[vh.CPath] && appendSubs {
			keys, err := c.List(KVMountPath + spath)
			if err != nil {
				log.WithFields(log.Fields{
					"spath":       spath,
					"pfclient":    c,
					"appendSubs":  appendSubs,
					"t":           t,
					"t[vh.CPath]": t[vh.CPath],
					"type":        "vh.CPath",
					"storesecret": s,
					"error":       err}).Warn("got error while getting vault secret with client and spath for adding as subs to store secret. Continuing...")
			} else {
				for _, v := range keys {
					newsec := &Secret{
						Path: filepath.Join(spath, v),
						Mode: sfsfh.DIRREAD,
					}
					s.Subs = append(s.Subs, newsec)
				}
			}
		}
		return s, nil

	case t[vh.CKey]:
		content, err := vh.GetValueFromKey(c, KVMountPath+spath)
		if err != nil {
			return nil, err
		}
		return &Secret{
			Path:    spath,
			Mode:    sfsfh.FILEREAD,
			Content: content,
		}, nil

	default: // probably not enough permissions to determine type -> would probably be a directory
		return nil, fmt.Errorf("could not evaluate filetype of %s\n", spath)
	}
}

func (s *VaultKv) String() string {
	return "vault_kv"
}

// GetClient returns a postfinance vault client.
// The context is used to detect the calling user and loading his vault
// approleId
func GetClient(ctx context.Context) (*pfvault.Client, error) {
	// Get default vault client configuration
	conf := api.DefaultConfig()
	a := viper.GetString("store.vault.addr")
	conf.Address = a

	// check TLS settings
	if len(a) >= 5 && a[:5] == "https" {
		if err := configureTLS(conf); err != nil {
			log.WithFields(log.Fields{
				"address": a,
				"err":     err}).Fatal("got error while configuring TLS, shutting down")
		}
	}

	// Create new vault client with vault configuration
	vc, err := api.NewClient(conf) // VaultClient
	if err != nil {
		return nil, err
	}
	// Get user doing the filesystem request
	u, err := sfsfh.GetUserFromContext(ctx)
	if err != nil {
		return nil, err
	}
	// Read approleId from configfile
	approleId, err := getApproleId(u)
	if err != nil {
		return nil, err
	}
	// Login with approleId, get accessToken
	accessToken, err := VaultApproleLogin(vc, approleId)
	if err != nil {
		return nil, err
	}
	// Set accessToken and return pfvault Client
	vc.SetToken(accessToken)
	// Create new pfvault client with vault client
	pfc, err := pfvault.New(vc, KVMountPath) //PostFinanceClient
	if err != nil {
		log.WithFields(log.Fields{
			"clientconf":    conf,
			"user":          u,
			"pfvaultclient": pfc,
			"error":         err}).Error("got error while creating new pfc")
		return nil, err
	}
	if pfc == nil {
		log.WithFields(log.Fields{
			"clientconf":    conf,
			"user":          u,
			"pfvaultclient": pfc,
			"KVMountPath":   KVMountPath,
			"error":         err}).Error("pfc is unexpectedly nil, probably not enough permissions to list KVMountPath")
		return nil, fmt.Errorf("msg=\"pfc is unexpectedly nil, probably not enough permissions to list '%v' on vault server\"\n", KVMountPath)
	}
	log.WithFields(log.Fields{
		"clientconf":    conf,
		"user":          u,
		"pfvaultclient": pfc,
		"KVMountPath":   KVMountPath}).Debug("log values")
	return pfc, err
}

func configureTLS(c *api.Config) error {
	tls := api.TLSConfig{}
	if viper.IsSet("store.vault.tls.cacert") {
		tls.CACert = viper.GetString("store.vault.tls.cacert")
	}
	if viper.IsSet("store.vault.tls.capath") {
		tls.CAPath = viper.GetString("store.vault.tls.capath")
	}
	if viper.IsSet("store.vault.tls.clientcert") {
		tls.ClientCert = viper.GetString("store.vault.tls.clientcert")
	}
	if viper.IsSet("store.vault.tls.clientkey") {
		tls.ClientKey = viper.GetString("store.vault.tls.clientkey")
	}
	if viper.IsSet("store.vault.tls.tlsservername") {
		tls.TLSServerName = viper.GetString("store.vault.tls.tlsservername")
	}
	if viper.IsSet("store.vault.tls.insecure") {
		tls.Insecure = viper.GetBool("store.vault.tls.insecure")
	}
	err := c.ConfigureTLS(&tls)
	if c.Error != nil {
		return c.Error
	}
	return err
}

func getApproleId(u *user.User) (authToken string, err error) {
	spath := FinIdPath(u)
	log.WithFields(log.Fields{
		"username": u.Username,
		"spath":    spath}).Debug("log values")
	o, err := ioutil.ReadFile(spath)
	if err != nil {
		log.WithFields(log.Fields{
			"username": u.Username,
			"spath":    spath,
			"error":    err}).Error("could not read spath for getting approleId")
		return "", err
	}
	return strings.TrimSuffix(string(o), "\n"), nil
}

func FinIdPath(u *user.User) (spath string) {
	spath = viper.GetString("store.vault.roleid.file")
	overriddenusers := viper.GetStringMapString("store.vault.roleid.useroverride")
	log.WithFields(log.Fields{
		"user":           u,
		"username":       u.Name,
		"overridenusers": overriddenusers}).Debug("log values")
	if newpath, ok := overriddenusers[u.Username]; ok {
		spath = newpath
	}
	return strings.Replace(spath, "$HOME", u.HomeDir, 1)
}

func VaultApproleLogin(c *api.Client, approleId string) (accessToken string, err error) {
	data := map[string]interface{}{
		"role_id": approleId,
	}
	resp, err := c.Logical().Write("auth/approle/login", data)
	if err != nil {
		return "", err
	}
	if resp.Auth == nil {
		return "", errors.New("no auth info returned")
	}
	return resp.Auth.ClientToken, nil
}

func init() {
	v := VaultKv{}
	RegisterStore(&v)
}
