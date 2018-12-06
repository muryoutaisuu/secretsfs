package fio

import (
	"github.com/hanwen/go-fuse/fuse"
	"github.com/hanwen/go-fuse/fuse/nodefs"
)

// FIOProvider interface provides all necessary calls used by FUSE.
// FIOProvider implementations will be called by the secretsfs high-top filesytem.
//
// If you want to program your own FIO plugin, please implement this interface
// and register your provider with the RegisterProvider(*FIOMap) function.
type FIOProvider interface {
	GetAttr(string, *fuse.Context) (*fuse.Attr, fuse.Status)
	Open(string, uint32, *fuse.Context) (nodefs.File, fuse.Status)
	OpenDir(string, *fuse.Context) ([]fuse.DirEntry, fuse.Status)
	FIOPath() string
}

// FIOMap maps the FIOProvider to a MountPath.
// Used for registering FIOProviders.
type FIOMap struct {
	Provider FIOProvider
	Enabled bool
}
