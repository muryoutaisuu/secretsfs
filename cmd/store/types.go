package store

// Store interface provides all necessary calls to backend store
type Store interface {
	List(string) error
	Read(string) error
	Write(string, string) error
	Delete(string) error
	String() (string, error)
}
