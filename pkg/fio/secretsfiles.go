package fio

import (
	"github.com/hanwen/go-fuse/fuse"
	"github.com/hanwen/go-fuse/fuse/nodefs"
)

// FIOSecretsfiles is a Filesystem implementing the FIOPlugin interface that
// outputs secrets directly when doing a command like:
//	cat <mountpath>/secretsfiles/<secretsItem>
type FIOSecretsfiles struct{}

// GetAttr references to fuse.GetAttr
func (t *FIOSecretsfiles) GetAttr(name string, context *fuse.Context) (*fuse.Attr, fuse.Status) {
	return sto.GetAttr(name, context)
}

// OpenDir references to fuse.OpenDir
func (t *FIOSecretsfiles) OpenDir(name string, context *fuse.Context) ([]fuse.DirEntry, fuse.Status) {
	return sto.OpenDir(name, context)
}

// Open references to fuse.Open
func (t *FIOSecretsfiles) Open(name string, flags uint32, context *fuse.Context) (nodefs.File, fuse.Status) {
	content, status := sto.Open(name, flags, context)
	if status == fuse.OK && content != "" {
		return nodefs.NewDataFile([]byte(content)), status
	}
	return nil, status
}

// FIOPath returns name of implemented FIO plugin
func (t *FIOSecretsfiles) FIOPath() string {
	return "secretsfiles"
}

func init() {
	fioprov := FIOSecretsfiles{}
	fm := FIOMap{
		Provider: &fioprov,
	}
	RegisterProvider(&fm)
}
