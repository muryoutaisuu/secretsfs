// Secretsfs
//
// Secretsfs contains the high-top filesystem, that controls top-paths and
// correctly redirects calls to the correct FUSE Input/Output (FIO) plugin
//
// If the tool secretsfs is mounted at /mnt/secretsfs, then for /mnt/secretsfs/x
// x will be the paths of the registered FIO Plugins.
// like e.g.:
//	/mnt/secretsfs/secretsfiles
//	/mnt/secretsfs/templatefiles
//
// FUSE Calls will be redirected 1:1 to those plugins, only changes made are:
//	1. calls directly on the top layer, like `ls -ld /mnt/secretsfs/secretsfiles`
//	   that call will be answered by secretsfs itself
//	2. the called path will be shortened, so it matches more accurately
//	   that means, when a user calls `ls -la /mnt/secretsfs/secretsfiles/foo/bar`
//	   instead of passing the values /mnt/secretsfs/secretsfiles/foo/bar or secretsfiles/foo/bar
//	   the value foo/bar will be returned
// inspired by this example: https://github.com/hanwen/go-fuse/blob/master/example/hello/main.go
package secretsfs


import (
	"errors"
	"strings"
	"path/filepath"
	"os/user"

	"github.com/hanwen/go-fuse/fuse"
	"github.com/hanwen/go-fuse/fuse/nodefs"
	"github.com/hanwen/go-fuse/fuse/pathfs"
	log "github.com/sirupsen/logrus"

	"github.com/muryoutaisuu/secretsfs/pkg/fio"
	"github.com/muryoutaisuu/secretsfs/pkg/store"
	"github.com/muryoutaisuu/secretsfs/pkg/sfslog"
	sfsh "github.com/muryoutaisuu/secretsfs/pkg/sfshelpers"
)

// logging
var logger = log.NewEntry(log.StandardLogger())

// SecretsFS is the high-top filesystem.
// It contains references to FIOMap (mapping mountpath to a plugin) and the
// currently used store. 
// 
type SecretsFS struct {
	pathfs.FileSystem
	fms map[string]*fio.FIOMap
	store store.Store
}

// NewSecretsFS return a fully configured SecretsFS, that is ready to be mounted.
// Also does a pre-check whether a store was defined. Returns an error if that
// is not the case.
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
	logger.WithFields(log.Fields{"name":name, "context":context}).Info("log values")
	root, subpath := rootName(name)
	u,err := sfsh.GetUser(context)
	if err != nil {
		logger.WithFields(log.Fields{"name":name, "context":context, "error":err}).Info("got error while getting user information")
		return nil, fuse.EPERM
	}
	logger = defaultEntry(name, u, root, subpath)
	logger.Info("calling operation")

	if root == "" && subpath == "" {
		logger.Info("successfully delivered attributes")
		return &fuse.Attr{Mode: fuse.S_IFDIR | 0755,}, fuse.OK
	}
	if _,ok := sfs.fms[root]; ok && sfs.fms[root].Enabled {
		logger.Info("successfully delivered attributes")
		return sfs.fms[root].Provider.GetAttr(subpath, context)
	}
	logger.WithFields(log.Fields{"name":name, "context":context}).Info("no element found")
	return &fuse.Attr{}, fuse.ENOENT
}

func (sfs *SecretsFS) OpenDir(name string, context *fuse.Context) (c []fuse.DirEntry, code fuse.Status) {
	logger.WithFields(log.Fields{"name":name, "context":context}).Info("log values")
	root, subpath := rootName(name)
	u,err := sfsh.GetUser(context)
	if err != nil {
		logger.WithFields(log.Fields{"name":name, "context":context, "error":err}).Info("got error while getting user information")
		return nil, fuse.EPERM
	}
	logger = defaultEntry(name, u, root, subpath)
	logger.Info("calling operation")

	if name == "" {
		c = []fuse.DirEntry{}
		for k := range sfs.fms {
			c = append(c, fuse.DirEntry{Name: k, Mode: fuse.S_IFDIR})
		}
		logger.Info("successfully listed directory")
		return c, fuse.OK
	}
	if _,ok := sfs.fms[root]; ok && sfs.fms[root].Enabled {
		logger.Info("successfully listed directory")
		return sfs.fms[root].Provider.OpenDir(subpath, context)
	}
	logger.WithFields(log.Fields{"name":name, "context":context}).Info("no element found")
	return nil, fuse.ENOENT
}

func (sfs *SecretsFS) Open(name string, flags uint32, context *fuse.Context) (file nodefs.File, code fuse.Status) {
	logger.WithFields(log.Fields{"name":name, "context":context}).Info("log values")
	root, subpath := rootName(name)
	u,err := sfsh.GetUser(context)
	if err != nil {
		logger.WithFields(log.Fields{"name":name, "context":context, "error":err}).Info("got error while getting user information")
		return nil, fuse.EPERM
	}
	logger = defaultEntry(name, u, root, subpath)
	logger.Info("calling operation")

	if name == "" {
		logger.WithFields(log.Fields{"name":name, "context":context}).Info("no element found")
		return nil, fuse.EINVAL
	}
	if _,ok := sfs.fms[root]; ok && sfs.fms[root].Enabled {
		logger.Info("successfully delivered file")
		return sfs.fms[root].Provider.Open(subpath, flags, context)
	}
	logger.WithFields(log.Fields{"name":name, "context":context}).Info("no element found")
	return nil, fuse.EPERM
}


func defaultEntry(name string, user *user.User, root, subpath string) *log.Entry {
	return sfslog.DefaultEntry(name, user).WithFields(log.Fields{
    "root": root,
    "subpath": subpath,
  })
}



// rootName calculates, which FIO the call came from and what the subpath for
// the FIO is
func rootName(path string) (root, subpath string) {
  list := strings.Split(path, string(filepath.Separator))
  root = list[0]
  subpath = filepath.Join(list[1:]...)
  return
}
