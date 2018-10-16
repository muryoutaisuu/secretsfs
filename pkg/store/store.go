package store

// stores contains all registered Stores
var store Store

// Store returns currently active Store Implementation
func GetStore() Store {
	return store
}

// RegisterStore function registers Stores.
// It is called from within the init() Function of other Store Implementations.
func RegisterStore(s Store) {
	store = s
}

