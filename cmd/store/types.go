package store

import (
	"os/user"
)

// Store interface provides all necessary calls to backend store
type Store interface {
	List(*user.User, string) error
	Read(*user.User, string) error
	Write(*user.User, string, string) error
	Delete(*user.User, string) error
	String() (string, error)
}
