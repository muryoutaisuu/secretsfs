package secretsfs

// after the example: https://github.com/hanwen/go-fuse/blob/master/example/hello/main.go

import (
	"errors"
	"strings"
	"path/filepath"

	"github.com/hanwen/go-fuse/fuse"
	"github.com/hanwen/go-fuse/fuse/nodefs"
	"github.com/hanwen/go-fuse/fuse/pathfs"

	"github.com/Muryoutaisuu/secretsfs/pkg/fio"
	"github.com/Muryoutaisuu/secretsfs/pkg/store"
	"github.com/Muryoutaisuu/secretsfs/pkg/sfslog"
)

var Log *sfslog.Log = sfslog.Logger()

type SecretsFS struct {
	pathfs.FileSystem
	fms map[string]*fio.FIOMap
	store store.Store
}

func NewSecretsFS(fs pathfs.FileSystem, fms map[string]*fio.FIOMap, s store.Store) (*SecretsFS, error) {
	if s == nil {
		return nil, errors.New("could not initialize store, store is nil!")
	}
	sfs := SecretsFS{
		FileSystem: fs,
		fms: fms,
		store: s,
	}
	return &sfs, nil
}

func (sfs *SecretsFS) GetAttr(name string, context *fuse.Context) (*fuse.Attr, fuse.Status) {
	root, subpath := rootName(name)
	Log.Debug.Printf("ops=GetAttr name=\"%v\" root=\"%v\" subpath=\"%v\"\n",name,root,subpath)
	if root == "" && subpath == "" {
		return &fuse.Attr{Mode: fuse.S_IFDIR | 0755,}, fuse.OK
	}
	if _,ok := sfs.fms[root]; ok {
		return sfs.fms[root].Provider.GetAttr(subpath, context)
	}
	return &fuse.Attr{}, fuse.ENOENT
}

func (sfs *SecretsFS) OpenDir(name string, context *fuse.Context) (c []fuse.DirEntry, code fuse.Status) {
	root, subpath := rootName(name)
	Log.Debug.Printf("ops=GetAttr name=\"%v\" root=\"%v\" subpath=\"%v\"\n",name,root,subpath)
	if name == "" {
		c = []fuse.DirEntry{}
		for k := range sfs.fms {
			c = append(c, fuse.DirEntry{Name: k, Mode: fuse.S_IFDIR})
		}
		return c, fuse.OK
	}
	if _,ok := sfs.fms[root]; ok {
		return sfs.fms[root].Provider.OpenDir(subpath, context)
	}
	return nil, fuse.ENOENT
}

func (sfs *SecretsFS) Open(name string, flags uint32, context *fuse.Context) (file nodefs.File, code fuse.Status) {
	root, subpath := rootName(name)
	Log.Debug.Printf("ops=GetAttr name=\"%v\" root=\"%v\" subpath=\"%v\"\n",name,root,subpath)
	if name == "" {
		return nil, fuse.EINVAL
	}
	if _,ok := sfs.fms[root]; ok {
		return sfs.fms[root].Provider.Open(subpath, flags, context)
	}
	return nil, fuse.EPERM
}



func rootName(path string) (root, subpath string) {
  list := strings.Split(path, string(filepath.Separator))
  root = list[0]
  subpath = filepath.Join(list[1:]...)
  return
}

