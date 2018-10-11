package secretsfs

// taken from example https://github.com/hanwen/go-fuse/blob/master/example/hello/main.go


import (
	"github.com/hanwen/go-fuse/fuse/pathfs"
)


type SecretsFS struct {
	pathfs.FileSystem
}

var _ pathfs.FileSystem = (*SecretsFS)(nil)
