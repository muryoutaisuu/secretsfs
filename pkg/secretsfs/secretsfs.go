package secretsfs

// after the example: https://github.com/hanwen/go-fuse/blob/master/example/hello/main.go

import (
	"log"

	"github.com/hanwen/go-fuse/fuse"
	"github.com/hanwen/go-fuse/fuse/nodefs"
	"github.com/hanwen/go-fuse/fuse/pathfs"
	"github.com/Muryoutaisuu/secretsfs/pkg/fio"
)

type SecretsFS struct {
	pathfs.FileSystem
	fms map[string]*fio.FIOMap
}

func NewSecretsFS(fs pathfs.FileSystem, fms map[string]*fio.FIOMap) (*SecretsFS, error) {
	sfs := SecretsFS{
		FileSystem: fs,
		fms: fms,
	}
	return &sfs, nil
}

func (sfs *SecretsFS) GetAttr(name string, context *fuse.Context) (*fuse.Attr, fuse.Status) {
	_,ok := sfs.fms[name]
	switch ok {
		case true : {
			return &fuse.Attr{
				Mode: fuse.S_IFDIR | 0755,
			}, fuse.OK
		}
	}
	switch name {
	case "file.txt":
		return &fuse.Attr{
			Mode: fuse.S_IFREG | 0644, Size: uint64(len(name)),
		}, fuse.OK
	case "natiitest.txt":
		return &fuse.Attr{
			Mode: fuse.S_IFREG | 0644, Size: uint64(len(name)),
		}, fuse.OK
	case "":
		return &fuse.Attr{
			Mode: fuse.S_IFDIR | 0755,
		}, fuse.OK
	}
	log.Fatal(name +" does not exist")
	return nil, fuse.ENOENT
}

func (sfs *SecretsFS) OpenDir(name string, context *fuse.Context) (c []fuse.DirEntry, code fuse.Status) {
	if _,ok := sfs.fms[name]; ok {
		c = []fuse.DirEntry{{Name: "natiitest.txt", Mode: fuse.S_IFREG}}
		return c, fuse.OK
	}
	if name == "" {
		c = []fuse.DirEntry{}
		for k := range sfs.fms {
			c = append(c, fuse.DirEntry{Name: k, Mode: fuse.S_IFDIR})
		}
		return c, fuse.OK
	}
	return nil, fuse.ENOENT
}

func (sfs *SecretsFS) Open(name string, flags uint32, context *fuse.Context) (file nodefs.File, code fuse.Status) {
	if name != "file.txt" {
		return nil, fuse.ENOENT
	}
	if flags&fuse.O_ANYWRITE != 0 {
		return nil, fuse.EPERM
	}
	return nodefs.NewDataFile([]byte(name)), fuse.OK
}
