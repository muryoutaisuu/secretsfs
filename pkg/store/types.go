package store

import (
	"github.com/hanwen/go-fuse/fuse"
	//"github.com/hanwen/go-fuse/fuse/nodefs"
)

// Store interface provides all necessary calls to backend store
type Store interface {
	GetAttr(name string, context *fuse.Context) (*fuse.Attr, fuse.Status)
	Open(name string, flags uint32, context *fuse.Context) (string, fuse.Status)
	OpenDir(name string, context *fuse.Context) ([]fuse.DirEntry, fuse.Status)

	// String returns a string containing the name of the store plugin.
  // It decides, how to react on FUSE-calls.
  // String was done as a function rather than as an attribute, because it is not
  // possible to define attributes for interfaces in golang.
  // Also see https://github.com/golang/go/issues/23796
	String() (string)
}
