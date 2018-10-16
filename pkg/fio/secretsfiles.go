package fio

import (
	"github.com/hanwen/go-fuse/fuse"
	"github.com/hanwen/go-fuse/fuse/nodefs"

	"github.com/Muryoutaisuu/secretsfs/pkg/store"
)

type FIOSecretsfiles struct {}

var sto store.Store

func (t *FIOSecretsfiles) GetAttr(name string, context *fuse.Context) (*fuse.Attr, fuse.Status) {
	return sto.GetAttr(name, context)
}

func (t *FIOSecretsfiles) OpenDir(name string, context *fuse.Context) ([]fuse.DirEntry, fuse.Status) {
	return sto.OpenDir(name, context)
}

func (t *FIOSecretsfiles) Open(name string, flags uint32, context *fuse.Context) (nodefs.File, fuse.Status) {
	return sto.Open(name, flags, context)
}




func init() {
       fm := FIOMap {
             MountPath: "secretsfiles",
             Provider: &FIOSecretsfiles{},
       }
			 sto = store.GetStore()
       
       RegisterProvider(&fm)
}
