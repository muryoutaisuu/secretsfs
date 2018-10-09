package store

// stores contains all registered Stores
var stores []Store

// RegisterStore registers Stores
func RegisterStore(s Store) {
       stores = append(stores, s)
}

