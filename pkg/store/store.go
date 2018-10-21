package store

import (
	"github.com/Muryoutaisuu/secretsfs/pkg/sfslog"
)

// stores contains all registered Stores
var store Store

// Log contains all the needed Loggers
var Log *sfslog.Log = sfslog.Logger()

// Store returns currently active Store Implementation
func GetStore() Store {
	return store
}

// RegisterStore function registers Stores.
// It is called from within the init() Function of other Store Implementations.
func RegisterStore(s Store) {
	store = s
}

