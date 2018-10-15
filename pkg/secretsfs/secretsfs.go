package secretsfs

// after the example: https://github.com/hanwen/go-fuse/blob/master/zipfs/memtree.go

import (
	//"github.com/hanwen/go-fuse/fuse"
	//"github.com/hanwen/go-fuse/fuse/nodefs"
	"github.com/hanwen/go-fuse/fuse/pathfs"
	"github.com/Muryoutaisuu/secretsfs/pkg/fio"
)

type SecretsFS struct {
	pathfs.FileSystem
	fms map[string]*fio.FIOMap
}
