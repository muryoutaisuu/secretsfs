package secretsfs

import (
	"context"
	"syscall"

	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	"github.com/muryoutaisuu/secretsfs/pkg/store"
)

// sto contains the currently set store
var sto *store.Store

// fiomaps contains all FIOMaps, that map FIORoot to MountPaths
var fiomaps map[string]*FIOMap = make(map[string]*FIOMap)

// enabledFIOs contains all enabled fios due to configuration
var enabledFIOs []string

// FIORoot interface describes functions a new FIO plugin should implement.
type FIORoot interface {
	// Node Operations
	Readdir(n *SfsNode, ctx context.Context) (out fs.DirStream, errno syscall.Errno)
	Lookup(n *SfsNode, ctx context.Context, name string, out *fuse.EntryOut) (node *fs.Inode, errno syscall.Errno)
	Open(n *SfsNode, ctx context.Context, flags uint32) (fh fs.FileHandle, fuseFlags uint32, errno syscall.Errno)
	Read(n *SfsNode, ctx context.Context, f fs.FileHandle, dest []byte, off int64) (fuse.ReadResult, syscall.Errno)
	Getattr(n *SfsNode, ctx context.Context, fh fs.FileHandle, out *fuse.AttrOut) syscall.Errno

	// FIOPath() is used for registering and finding FIOMaps
	FIOPath() string
}

// FIOMap maps the FIORoot Node to a Mountpath
// Used for registering FIORoots to the secretsfs rootnode
type FIOMap struct {
	Root    FIORoot
	Enabled bool
}

// RegisterRoot registers FIOMaps.
// To be used inside of init() function of plugins.
// If FIO is enabled due to configuration, also enable it.
// If not, only register in in Disabled state.
func RegisterRoot(fm *FIOMap) {
	for _, enabledfio := range enabledFIOs {
		if fm.Root.FIOPath() == enabledfio {
			fm.Enabled = true
		}
	}
	fiomaps[fm.Root.FIOPath()] = fm
}

// FIOMaps returns all registered FIO Plugins mapped with their mountpaths.
// Generally used by rootnode so it can correctly redirect
// calls to corresponding plugins.
func FIOMaps() map[string]*FIOMap {
	return fiomaps
}

// getFIORootFromRootPath returns the FIORoot if rootpath is in fms
func getFIORootFromRootPath(rootpath string) FIORoot {
	log.WithFields(log.Fields{"rootpath": rootpath, "fiomaps": fiomaps}).Debug("log values")
	for rpath, fm := range fiomaps {
		if rpath == rootpath {
			return fm.Root
		}
	}
	return nil
}

// RootPathsEnabled returns all registered fioRootPaths that are enabled
func RootPathsEnabled() []string {
	enabledRoots := []string{}
	for _, v := range fiomaps {
		if v.Enabled {
			enabledRoots = append(enabledRoots, v.Root.FIOPath())
		}
	}
	return enabledRoots
}

// IsRootPath checks whether the given rootpath is registered and enabled in fiomaps
func IsRootPath(rootpath string) bool {
	if rootpath[:1] == "/" {
		rootpath = rootpath[1:]
	}
	for k, v := range fiomaps {
		if rootpath == k {
			return v.Enabled
		}
	}
	return false
}

// FIOMapsEnabled returns a map[string]*FIOMap only with enabled FIOMaps
func FIOMapsEnabled() map[string]*FIOMap {
	enabledFIOMaps := make(map[string]*FIOMap)
	for k, v := range fiomaps {
		if v.Enabled {
			enabledFIOMaps[k] = v
		}
	}
	return enabledFIOMaps
}

func loadEnabledFIOs(init bool) {
	enabledFIOs = viper.GetStringSlice("fio.enabled")
	// if not from init, then reload all fiomaps
	if !init {
		for _, fm := range fiomaps {
			RegisterRoot(fm)
		}
	}
}

func init() {
	sto = store.GetStore()
	loadEnabledFIOs(true)
}
