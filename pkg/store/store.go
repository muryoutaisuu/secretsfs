package store

import (
	"github.com/muryoutaisuu/secretsfs/pkg/sfslog"
	"github.com/spf13/viper"
	//"github.com/muryoutaisuu/secretsfs/cmd/secretsfs/config"
)

// store contains the registered Store
var store Store

// available stores
var stores []string

// Log contains all the needed Loggers
var Log *sfslog.Log = sfslog.Logger()

// Store returns currently active Store Implementation
func GetStore() Store {
	return store
}

// RegisterStore registers available stores
// if a store is also set to be the backend store it will be set here
func RegisterStore(s Store) {
	stores = append(stores, s.String())
	if viper.GetString("CURRENT_STORE") == s.String() {
		store = s
	}
}

// GetStores returns all registered stores.
// Registered stores are all available stores that a user may configure as a
// store of secretsfs.
func GetStores() []string {
	return stores
}




func init() {
	stores = []string{}
}
