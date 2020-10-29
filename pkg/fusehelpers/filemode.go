package fusehelpers

import (
	"github.com/hanwen/go-fuse/v2/fuse"
)

const (
	FILEREAD   = fuse.S_IFREG + 0x0755
	FILENOREAD = fuse.S_IFREG + 0x0700
	DIRREAD    = fuse.S_IFDIR + 0x0755
	DIRNOREAD  = fuse.S_IFDIR + 0x0700
)

func IsFile(mode int64) bool {
	return fuse.S_IFREG&mode == fuse.S_IFREG
}

func IsDir(mode int64) bool {
	return fuse.S_IFDIR&mode == fuse.S_IFDIR
}
