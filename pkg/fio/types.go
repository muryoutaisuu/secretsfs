package fio

import (
	"github.com/hanwen/go-fuse/fuse"
	"github.com/hanwen/go-fuse/fuse/nodefs"
)

// FIOProvider interface provides all necessary calls used by FUSE
type FIOProvider interface {
	GetAttr(string, *fuse.Context) (*fuse.Attr, fuse.Status)
	// uint32 = flags
	Open(string, uint32, *fuse.Context) (nodefs.File, fuse.Status)
	OpenDir(string, *fuse.Context) ([]fuse.DirEntry, fuse.Status)
}

// FIOMap maps the FIOProvider to a MountPath
type FIOMap struct {
	MountPath string
	Provider FIOProvider
}
