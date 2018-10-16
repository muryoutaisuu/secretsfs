package store

import (
	"github.com/hanwen/go-fuse/fuse"
	"github.com/hanwen/go-fuse/fuse/nodefs"
)

// Store interface provides all necessary calls to backend store
type Store interface {
	GetAttr(name string, context *fuse.Context) (*fuse.Attr, fuse.Status)
	Open(name string, flags uint32, context *fuse.Context) (nodefs.File, fuse.Status)
	OpenDir(name string, context *fuse.Context) ([]fuse.DirEntry, fuse.Status)
	//List(*user.User, string) error
	//Read(*user.User, string) error
	//Write(*user.User, string, string) error
	//Delete(*user.User, string) error
	String() (string, error)
}
