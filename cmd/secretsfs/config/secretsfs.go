package config

import (
	"fmt"
	"bytes"

	"github.com/spf13/viper"
)

// Default configurations
var configDefaults = []byte(`
---
fio:
  templatefiles:
    PATH_TO_TEMPLATES: /etc/secretsfs/templates/
store:
  vault:
    FILE_ROLEID: .vault-roleid
    VAULT_ADDR: http://127.0.0.1:8200
    # taken from https://www.vaultproject.io/api/secret/kv/kv-v2.html
    MTDATA: secret/metadata/
    DTDATA: secret/data/
`)

// https://github.com/spf13/viper#reading-config-files
func InitConfig() {
	viper.SetConfigName("secretsfs")
	viper.AddConfigPath("/etc/secretsfs/")
	viper.AddConfigPath("$HOME/.secretsfs")  // call multiple times to add many search paths
	err := viper.ReadInConfig() // Find and read the config file
	if err != nil { // Handle errors reading the config file
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}

	// read automatically all envs with Prefix SFS_
	viper.SetEnvPrefix("SFS")
	viper.AutomaticEnv()
	// also read vault addr env
	// needs both parameters, else prefix would be prefixed
	viper.BindEnv("VAULT_ADDR","VAULT_ADDR")

	viper.SetConfigType("yaml")
	viper.ReadConfig(bytes.NewBuffer(configDefaults))
}

func GetConfigDefaults() *[]byte {
	return &configDefaults
}

func init() {
	InitConfig()
}
