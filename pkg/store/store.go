package store

import (
	"os/user"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	"github.com/muryoutaisuu/secretsfs/pkg/sfslog"
)

// store contains the registered Store
var store Store

// available stores
var stores []string

// logging
var logger = log.NewEntry(log.StandardLogger())

// GetStore returns currently active Store Implementation
func GetStore() Store {
	return store
}

// RegisterStore registers available stores
// if a store is also set to be the backend store it will be set here
func RegisterStore(s Store) {
	stores = append(stores, s.String())
	if viper.GetString("store.enabled") == s.String() {
		store = s
	}
}

// GetStores returns all registered stores.
// Registered stores are all available stores that a user may configure as a
// store of secretsfs.
func GetStores() []string {
	return stores
}

func defaultEntry(name string, user *user.User) *log.Entry {
	return sfslog.DefaultEntry(name, user)
}

func init() {
	stores = []string{}
}
