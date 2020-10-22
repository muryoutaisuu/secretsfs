package store

import (
	"context"
	//"github.com/hanwen/go-fuse/v2/fs"
	//"github.com/hanwen/go-fuse/v2/fuse"
)

// store contains the registered Store
var store Store

// available stores
var stores []string

// GetStore returns currently active Store Implementation
func GetStore() *Store {
	return &store
}

// GetStores returns all registered stores.
// Registered stores are all available stores that a user may configure as a
// store of secretsfs.
func GetStores() []string {
	return stores
}

// RegisterStore registers available stores
// if a store is also set to be the backend store it will be set here
func RegisterStore(s Store) {
	stores = append(stores, s.String())
	if s.String() == "vault_kv" {
		store = s
	}
	//if viper.GetString("store.enabled") == s.String() {
	//	store = s
	//}
}

// Store interface describes functions a new store should implement.
type Store interface {
	// for convenience
	GetSecret(spath string, ctx context.Context) (secret *Secret, err error)

	// String() is used to distinguish between different store implementations
	String() string
}

func init() {
	stores = []string{}
}
