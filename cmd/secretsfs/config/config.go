// config contains the config information about secretsfs.
// it contains the default configuration, so that it can be accessed and set
// without any worries.
package config

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

// configDefaults contains the default configurations.
// Those will be set on startup, if not overwritten via environment variables
// or a userdefined configurationfile.
var configDefaults = []byte(`
---
# General
general:
  configuration:
    paths:
    #- /etc/secretsfs/
    #- $HOME/.secretsfs
    #configfile: secretsfs  # without file type

  # logging levels may be: {trace,debug,info,warn,error,fatal,panic}
  logging:
    level: info

fio:
  enabled:
    - secretsfiles
    - templatefiles
    - internal
  templatefiles:
    # add additional locations for template files
    # the files in '/etc/secretsfs/templates/' for example will be mapped to
    # 'templatefiles/default/'
    templatespaths:
      default: /etc/secretsfs/templates/
      #applA: /appl/applA
  secretsfiles:
  internal:
    # privileges given to users or groups for listing and reading files in internal
    # do not make this readable for all, as it may contain critical data due to path namings
    privileges:
      users:
        - root
      groups:
        - admin

store:
  enabled: vault
  vault:
    roleid:
      # path configuration defines, where to look for the vault roleid token
      # $HOME will be substituted with the user's corresponding home directory
      # according to variable HomeDir in https://golang.org/pkg/os/user/#User
      # it *MUST* be uppcerase
      file: "$HOME/.vault-roleid"

      # useroverride configures paths per user, may be used to overwrite default
      # store.vault.roleid.file for some users
      # takes precedence over store.vault.roleid.file
      # store.vault.roleid.useroverride will *NOT* fallback to store.vault.roleid.file
      #useroverride:
      #  <usernameA>: <path>

    # address of the vault instance, that shall be accessed
    # differenciates between http:// and https:// protocols
    # defaults to a local dev instance
    addr: http://127.0.0.1:8200

    # vault TLS Configurations
    # for more information, see https://pkg.go.dev/github.com/hashicorp/vault/api#TLSConfig
    tls:
      #cacert: <path to PEM-encoded CA file>
      #capath: <path to directory of PEM-encoded CA files>
      #clientcert: <path to certificate for backend communication>
      #clientkey: <path to private key for backend communication>
      #tlsservername: <used for setting SNI host>
      #insecure: <disable TLS verification>
`)

// InitConfig reads all configurations and sets them.
// Order is (first match counts):
//	1. Environment variables
//	2. Configurationfile $HOME/.secretsfs/secretsfs.yaml
//	3. Configurationfile provided by environment variable SFS_CONFIG_FILE
//	4. Configurationfile /etc/secretsfs/secretsfs.yaml
//	5. Hardcoded configurations from var configDefaults
// This function is executed in init().
//
// https://github.com/spf13/viper#reading-config-files
func InitConfig() {
	// read defaults first
	viper.SetConfigType("yaml")
	viper.ReadConfig(bytes.NewBuffer(configDefaults))

	// read automatically all envs with Prefix SFS_
	viper.SetEnvPrefix("SFS")
	viper.AutomaticEnv()

	// also read vault addr env
	// needs both parameters for BindEnv, else prefix would be prefixed
	viper.BindEnv("store.vault.addr", "VAULT_ADDR")

	// read config file specific things first and overwrite if necessary
	viper.SetConfigName("secretsfs")
	viper.AddConfigPath("$HOME/.secretsfs") // call multiple times to add many search paths
	if viper.IsSet("general.configuration.configfile") {
		viper.SetConfigName(viper.GetString("general.configuration.configfile"))
	}

	// add config paths of ENV var first so it overwrites any other config?
	// TODO: check, whether it really works like this
	viper.AddConfigPath("/etc/secretsfs/")
	if viper.IsSet("general.configuration.paths") {
		paths := viper.GetStringSlice("general.configuration.paths")
		for _, p := range paths {
			viper.AddConfigPath(p)
		}
	}

	// read configuration from config files
	err := viper.MergeInConfig()                                                                                       // Find and read the config files
	if err != nil && !strings.Contains(err.Error(), "Config File") && !strings.Contains(err.Error(), "Not Found in") { // Handle errors reading the config file
		panic(fmt.Errorf("%s\n", err))
	}
}

// GetConfigDefaults returns the Contents of configDefaults as *[]byte.
// If you need string, you can also call GetStringConfigDefaults().
func GetConfigDefaults() *[]byte {
	return &configDefaults
}

// GetStringConfigDefaults returns the Contents of configDefaults converted as string.
func GetStringConfigDefaults() string {
	return string(configDefaults)
}

func init() {
	InitConfig()
}
