package secretsfs

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/hanwen/go-fuse/v2/fuse"
)

// pathsInodes contains all registered inodes so far, mapped to their paths
var pathsInodes map[string]uint64 = make(map[string]uint64)

// GetInode returns a valid inode for npath. If it isn't registered yet, it will
// be registered
func GetInode(npath string) uint64 {
	// clip trailing '/'
	npath = trimPath(npath)
	// get inode if existent
	inode, ok := pathsInodes[npath]
	// if not existent, assign new inode
	if !ok {
		inode = uint64(len(pathsInodes) + 2)
		pathsInodes[npath] = inode
	}
	return inode
}

// GetInodeIfRegistered returns the inode if it is registered, won't register it
// if it isn't already registered
func GetInodeIfRegistered(npath string) (uint64, bool) {
	npath = trimPath(npath)
	inode, ok := pathsInodes[npath]
	return inode, ok
}

// trimPath removes '/' if it is the last character and returns resulting string
func trimPath(npath string) string {
	if npath[len(npath)-1:len(npath)] == "/" {
		npath = npath[:len(npath)-1]
	}
	return npath
}

// rootName calculates, which FIO the call came from and what the subpath for
// the FIO is
func rootName(npath string) (rootpath, subpath string) {
	if len(npath) == 0 || npath == "/" {
		return "", ""
	} else if npath[0:1] == "/" {
		npath = npath[1:len(npath)]
	}
	list := strings.Split(npath, string(filepath.Separator))
	rootpath = list[0]
	subpath = filepath.Join(list[1:]...)
	return
}

// GetMapStringKeys returns []string
// containing all keys from a map[string]interface{}
func GetMapStringKeys(m map[string]interface{}) []string {
	keys := make([]string, len(m))
	i := 0
	for k := range m {
		keys[i] = k
		i++
	}
	return keys
}

// getModeFromFileInfo returns the corresponding fuse.Attr.Mode of a os.FileInfo
func getModeFromFileInfo(fi os.FileInfo) uint32 {
	if fi.IsDir() {
		return fuse.S_IFDIR
	}
	return fuse.S_IFREG
}

// PrettyPrint variable (struct, map, array, slice) in Golang
// https://siongui.github.io/2016/01/30/go-pretty-print-variable/
func PrettyPrint(v interface{}) ([]byte, error) {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return nil, err
	}
	return b, nil
}
