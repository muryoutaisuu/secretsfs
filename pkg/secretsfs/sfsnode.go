package secretsfs

import (
	"context"
	"fmt"
	"syscall"

	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
	log "github.com/sirupsen/logrus"
)

type SfsNode struct {
	fs.Inode
	npath string // node path
	fms   map[string]*FIOMap
}

func NewNode(npath string) *SfsNode {
	return &SfsNode{
		npath: npath,
	}
}

func GetNewRootNode(npath string, fms map[string]*FIOMap) *SfsNode {
	_ = GetInode(npath)
	return &SfsNode{
		npath: npath,
		fms:   fms,
	}
}

func (n *SfsNode) root() *SfsNode {
	return n.Root().Operations().(*SfsNode)
}

func (n *SfsNode) NPath() string {
	return n.npath
}

// Readdir
var _ = (fs.NodeReaddirer)((*SfsNode)(nil))

func (n *SfsNode) Readdir(ctx context.Context) (fs.DirStream, syscall.Errno) {
	defer log.WithFields(log.Fields{"pathsInodes": pathsInodes}).Debug("log values")
	log.WithFields(log.Fields{
		"nType":   fmt.Sprintf("%T", n),
		"n":       n,
		"n.npath": n.npath}).Debug("log values")
	if n.npath == "/" {
		var rootnodes []fuse.DirEntry
		for _, rpath := range RootPathsEnabled() {
			ino := GetInode("/" + rpath)
			rootnode := fuse.DirEntry{
				Name: rpath,
				Ino:  ino,
				Mode: fuse.S_IFDIR,
			}
			rootnodes = append(rootnodes, rootnode)
		}
		return fs.NewListDirStream(rootnodes), fs.OK
	}

	rootpath, _ := rootName(n.npath)
	log.WithFields(log.Fields{"rootpath": rootpath}).Debug("log values")
	fr := getFIORootFromRootPath(rootpath)
	if fr == nil {
		log.Println("returning syscall.ENOENT")
		log.WithFields(log.Fields{
			"n":        n,
			"n.npath":  n.npath,
			"rootpath": rootpath,
			"fr":       fr,
			"calling":  "getFIORootFromRootPath(rootpath)"}).Error("could not retrieve FIORoot from rootpath")
		return nil, syscall.ENOENT
	}
	log.WithFields(log.Fields{
		"n":            n,
		"n.npath":      n.npath,
		"rootpath":     rootpath,
		"fr":           fr,
		"fr.FIOPath()": fr.FIOPath()}).Debug("log values")
	return fr.Readdir(n, ctx)
}

// Open File
// This is only used for creating a filehandle. Since I work without FileHandles,
// just return, and work then with the Read(...) Method
var _ = (fs.NodeOpener)((*SfsNode)(nil))

func (n *SfsNode) Open(ctx context.Context, flags uint32) (fh fs.FileHandle, fuseFlags uint32, errno syscall.Errno) {
	defer log.WithFields(log.Fields{"pathsInodes": pathsInodes}).Debug("log values")
	log.WithFields(log.Fields{"n": n, "n.npath": n.npath}).Debug("log values")
	rootpath, _ := rootName(n.npath)
	fr := getFIORootFromRootPath(rootpath)
	if fr != nil {
		log.WithFields(log.Fields{
			"n":            n,
			"n.npath":      n.npath,
			"rootpath":     rootpath,
			"fr":           fr,
			"fr.FIOPath()": fr.FIOPath()}).Debug("delegating Open to FIORoot")
		return fr.Open(n, ctx, flags)
	}
	return nil, 0, 0
}

// Read File
var _ = (fs.NodeReader)((*SfsNode)(nil))

func (n *SfsNode) Read(ctx context.Context, fh fs.FileHandle, dest []byte, off int64) (fuse.ReadResult, syscall.Errno) {
	defer log.WithFields(log.Fields{"pathsInodes": pathsInodes}).Debug("log values")
	log.WithFields(log.Fields{"n": n, "n.npath": n.npath}).Debug("log values")
	rootpath, _ := rootName(n.npath)
	fr := getFIORootFromRootPath(rootpath)
	if fr == nil {
		log.Println("returning syscall.ENOENT")
		log.WithFields(log.Fields{
			"n":        n,
			"n.npath":  n.npath,
			"rootpath": rootpath,
			"fr":       fr,
			"calling":  "getFIORootFromRootPath(rootpath)"}).Error("could not retrieve FIORoot from rootpath")
		return nil, syscall.ENOENT
	}
	log.WithFields(log.Fields{
		"n":            n,
		"n.npath":      n.npath,
		"rootpath":     rootpath,
		"fr":           fr,
		"fr.FIOPath()": fr.FIOPath(),
		"calling":      "getFIORootFromRootPath(rootpath)"}).Debug("log values")
	return fr.Read(n, ctx, fh, dest, off)
}

// Lookup Node
var _ = (fs.NodeLookuper)((*SfsNode)(nil))

func (n *SfsNode) Lookup(ctx context.Context, name string, out *fuse.EntryOut) (*fs.Inode, syscall.Errno) {
	defer log.WithFields(log.Fields{"pathsInodes": pathsInodes}).Debug("log values")
	log.WithFields(log.Fields{"n": n, "n.npath": n.npath, "name": name}).Debug("log values")

	// root nodes
	if n.npath == "/" {
		inode, ok := GetInodeIfRegistered(n.npath + name)
		isroot := IsRootPath(name)
		if ok && isroot {
			stable := fs.StableAttr{
				Mode: fuse.S_IFDIR,
				Ino:  inode,
			}
			operations := NewNode(n.npath + name)
			child := n.NewPersistentInode(ctx, operations, stable)
			log.WithFields(log.Fields{
				"n":          n,
				"n.npath":    n.npath,
				"name":       name,
				"inode":      inode,
				"stable":     stable,
				"operations": operations,
				"child":      child,
				"isroot":     isroot,
				"calling":    "IsRootPath(name)"}).Debug("log values")
			return child, fs.OK
		}
	}

	rootpath, subpath := rootName(n.npath)
	log.WithFields(log.Fields{
		"n":        n,
		"n.npath":  n.npath,
		"name":     name,
		"rootpath": rootpath,
		"subpath":  subpath,
		"calling":  "rootName(n.npath)"}).Debug("log values")
	fr := getFIORootFromRootPath(rootpath)
	if fr == nil {
		log.Println("returning syscall.ENOENT")
		log.WithFields(log.Fields{
			"n":        n,
			"n.npath":  n.npath,
			"name":     name,
			"rootpath": rootpath,
			"subpath":  subpath,
			"fr":       fr,
			"calling":  "getFIORootFromRootPath(rootpath)"}).Error("could not retrieve FIORoot from rootpath")
		return nil, syscall.ENOENT
	}
	log.WithFields(log.Fields{
		"n":            n,
		"n.npath":      n.npath,
		"name":         name,
		"rootpath":     rootpath,
		"subpath":      subpath,
		"fr":           fr,
		"fr.FIOPath()": fr.FIOPath()}).Debug("log values")
	return fr.Lookup(n, ctx, name, out)
}

// GetAttrer
var _ = (fs.NodeGetattrer)((*SfsNode)(nil))

func (n *SfsNode) Getattr(ctx context.Context, fh fs.FileHandle, out *fuse.AttrOut) syscall.Errno {
	defer log.WithFields(log.Fields{"pathsInodes": pathsInodes}).Debug("log values")
	log.WithFields(log.Fields{"n": n, "n.npath": n.npath}).Debug("log values")

	if n.npath == "/" { // root
		return fs.OK
	}
	rootpath, _ := rootName(n.npath)
	fr := getFIORootFromRootPath(rootpath)
	if fr == nil {
		log.WithFields(log.Fields{
			"n":        n,
			"n.npath":  n.npath,
			"rootpath": rootpath,
			"fr":       fr,
			"calling":  "getFIORootFromRootPath(rootpath)"}).Error("could not retrieve FIORoot from rootpath")
		return syscall.ENOENT
	}
	log.WithFields(log.Fields{
		"n":            n,
		"n.npath":      n.npath,
		"rootpath":     rootpath,
		"fr":           fr,
		"fr.FIOPath()": fr.FIOPath()}).Debug("log values")
	return fr.Getattr(n, ctx, fh, out)
}

// OnAdder
var _ = (fs.NodeOnAdder)((*SfsNode)(nil))

func (n *SfsNode) OnAdd(ctx context.Context) {
	defer log.WithFields(log.Fields{"pathsInodes": pathsInodes}).Debug("log values")
	log.WithFields(log.Fields{"n": n, "n.npath": n.npath}).Debug("log values")
	if n.fms == nil {
		log.Println("OnAdd, leaving")
		log.WithFields(log.Fields{
			"n":         n,
			"n.npath":   n.npath,
			"n.fms":     n.fms,
			"condition": "n.fms == nil"}).Debug("leaving OnAdd because of empty n.fms")
		return
	}
	// register all rootPaths in advance, make them persistent
	for _, rootpath := range RootPathsEnabled() {
		_ = GetInode("/" + rootpath)
		log.WithFields(log.Fields{
			"n":        n,
			"n.npath":  n.npath,
			"n.fms":    n.fms,
			"rootpath": rootpath,
			"calling":  "GetInode(\"/\" + rootpath)"}).Debug("registered rootpath with OnAdd function")
	}
}
