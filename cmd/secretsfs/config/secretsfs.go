// config contains the config information about secretsfs.
// it contains the default configuration, so that it can be accessed and set
// without any worries.
package config

import (
	"fmt"
	"bytes"
	"strings"

	"github.com/spf13/viper"
)

// configDefaults contains the default configurations.
// Those will be set on startup, if not overwritten via environment variables
// or a userdefined configurationfile.
var configDefaults = []byte(`
---
### GENERAL
# CONFIG_PATHS: 
# - /etc/secretsfs/
# - $HOME/.secretsfs
# CONFIG_FILE: secretsfs  # without file type

# HTTPS Configurations
# HTTPS_CACERT: <path to PEM-encoded CA file>
# HTTPS_CAPATH: <path to directory of PEM-encoded CA files>
# HTTPS_CLIENTCERT: <path to certificate for backend communication>
# HTTPS_CLIENTKEY: <path to private key for backend communication>
# HTTPS_TLSSERVERNAME: <used for setting SNI host>
# HTTPS_INSECURE: <disable TLS verification>


### FIO
ENABLED_FIOS:
- secretsfiles
- templatefiles

# templatefiles
PATH_TO_TEMPLATES: /etc/secretsfs/templates/


### STORE
CURRENT_STORE: Vault

# vault
# path configuration defines, where to look for the vault roleid token
# $HOME will be substituted with the user's corresponding home directory
# according to variable HomeDir in https://golang.org/pkg/os/user/#User
# old: FILE_ROLEID: .vault-roleid
FILE_ROLEID: "$HOMEDIR/.vault-roleid"

# FILE_ROLEID_USER configures paths per user, may be used to overwrite default
# FILE_ROLEID for some users
# takes precedence over FILE_ROLEID
# FILE_ROLEID_USER will *NOT* fallback to FILE_ROLEID
#FILE_ROLEID_USER:
#  <usernameA>: <path>
VAULT_ADDR: http://127.0.0.1:8200
# taken from https://www.vaultproject.io/api/secret/kv/kv-v2.html
MTDATA: secret/
DTDATA: secret/


# fuse does not allow the character '/' inside of names of directories or files
# in vault k=v pairs of one secret will be shown as files, where k is the name
# of the file and v the value. k may also include names with a '/'.
# Those slashes will be substituted with the following character
# may also use some special characters, e.g. '§' or '°'
subst_char: _

`)

// InitConfig reads all configurations and sets them.
// Order is (first match counts):
//	1. Environment variables
//	2. Configurationfile $HOME/.secretsfs/secretsfs.yaml
//	3. Configurationfile provided by environment variable SFS_CONFIG_FILE
//	4. Configurationfile /etc/secretsfs/secretsfs.yaml
//	5. Hardcoded configurations from variable configDefaults
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
	viper.BindEnv("VAULT_ADDR","VAULT_ADDR")


	// read config file specific things first and overwrite if necessary
	viper.SetConfigName("secretsfs")
	viper.AddConfigPath("$HOME/.secretsfs")  // call multiple times to add many search paths
	if viper.IsSet("CONFIG_FILE") {
		viper.SetConfigName(viper.GetString("CONFIG_FILE"))
	}

	//   add config paths of ENV var first so it overwrites any other config?
	//   TODO: check, whether it really works like this
	viper.AddConfigPath("/etc/secretsfs/")
	if viper.IsSet("CONFIG_PATHS") {
		paths := viper.GetStringSlice("CONFIG_PATHS")
		for _,p := range paths {
			viper.AddConfigPath(p)
		}
	}

	// read configuration from config files
	err := viper.MergeInConfig() // Find and read the config files
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
