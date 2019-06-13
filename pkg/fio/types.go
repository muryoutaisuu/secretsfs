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
// The FUSE-calls are adopted from https://godoc.org/github.com/hanwen/go-fuse/fuse/pathfs#FileSystem
type FIOProvider interface {
	GetAttr(string, *fuse.Context) (*fuse.Attr, fuse.Status)
	Open(string, uint32, *fuse.Context) (nodefs.File, fuse.Status)
	OpenDir(string, *fuse.Context) ([]fuse.DirEntry, fuse.Status)

	// FIOPath returns a string containing the name and path of the FIO plugin.
	// It decides, on which subdirectory it will be available after mounting.
	// FIOPath was done as a function rather than as an attribute, because it is not
	// possible to define attributes for interfaces in golang.
	// Also see https://github.com/golang/go/issues/23796
	FIOPath() string
}

// FIOMap maps the FIOProvider to a MountPath.
// Used for registering FIOProviders.
type FIOMap struct {
	Provider FIOProvider
	Enabled  bool
}
