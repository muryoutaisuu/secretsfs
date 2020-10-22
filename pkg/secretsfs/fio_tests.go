package secretsfs

import (
	"context"
	"path/filepath"
	"syscall"

	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
	log "github.com/sirupsen/logrus"
	//"github.com/muryoutaisuu/secretsfs/pkg/store"
)

// FIOTest shall be a FIO example and can be used for simple testing
// following directory structure is implemented:
//   	tests/
//   	├── test1.txt
//   	├── test2.txt
//   	└── testdir
//   	  	├── test3.txt
//   	  	├── test4.txt
//   	  	└── test5.txt

type testNode struct {
	path    string
	isfile  bool
	content []byte
}
type testNodes struct {
	nodes []*testNode
}

var testnodes = testNodes{
	[]*testNode{
		&testNode{"/tests", false, nil},
		&testNode{"/tests/testdir", false, nil},
		&testNode{"/tests/test1.txt", true, []byte("This is the content of the file /tests/test1.txt\n")},
		&testNode{"/tests/test2.txt", true, []byte("This is the content of the file /tests/test2.txt\n")},
		&testNode{"/tests/testdir/test3.txt", true, []byte("This is the content of the file /tests/testdir/test3.txt\n")},
		&testNode{"/tests/testdir/test4.txt", true, []byte("This is the content of the file /tests/testdir/test4.txt\n")},
		&testNode{"/tests/testdir/test5.txt", true, []byte("This is the content of the file /tests/testdir/test5.txt\n")},
	},
}

func (t *testNodes) print() {
	for _, v := range t.nodes {
		log.WithFields(log.Fields{"testnode": v}).Debug("log values")
	}
}
func (t *testNodes) isDir(npath string) bool {
	for _, v := range t.nodes {
		if !v.isfile && v.path == npath {
			return true
		}
	}
	return false
}
func (t *testNodes) isFile(npath string) bool {
	for _, v := range t.nodes {
		if v.isfile && v.path == npath {
			return true
		}
	}
	return false
}
func (t *testNodes) getTestNodeByPath(npath string) *testNode {
	log.WithFields(log.Fields{"npath": npath}).Debug("log values")
	for _, v := range t.nodes {
		if v.path == npath {
			return v
		}
	}
	return nil
}

// getDirEntries returns all direntries matching spath in entries
// npath:		"/tests"
// entries:	"/tests/dir1"
//					"/tests/file1"
//					"/tests/dir1/file2"
//					"/otherdir"
// returns: "/tests/dir1"
//					"/tests/file1"
// not tested against edge cases, like entry "/testsdir1"
// only used for fio_tests.go, do *NOT* use in your projects
func (t *testNodes) getDirEntries(npath string) (direntries []*testNode) {
	l := len(npath)
	for _, v := range t.nodes {
		if len(v.path) < l || v.path == npath {
			continue
		}
		if v.path[:l] == npath && npath == filepath.Dir(v.path) {
			direntries = append(direntries, v)
			log.WithFields(log.Fields{"v.path": v.path, "direntries": direntries}).Debug("log values")
		}
	}
	return direntries
}
func (t *testNode) getMode() uint32 {
	if t.isfile {
		return fuse.S_IFREG
	}
	return fuse.S_IFDIR
}

type FIOTest struct{}

var _ = (FIORoot)((*FIOTest)(nil))

//Readdirer
func (sf *FIOTest) Readdir(n *SfsNode, ctx context.Context) (out fs.DirStream, errno syscall.Errno) {
	log.WithFields(log.Fields{"n": n, "n.npath": n.npath}).Debug("log values")
	if !testnodes.isDir(n.npath) {
		log.WithFields(log.Fields{"n.npath": n.npath}).Error("node is not a directory")
		return nil, syscall.ENOENT
	}
	log.WithFields(log.Fields{"n.npath": n.npath, "condition": "testnodes.isDir(n.npath) == true"}).Debug("node is a directory")

	entries := testnodes.getDirEntries(n.npath)
	var direntries []fuse.DirEntry

	for _, v := range entries {
		direntries = append(direntries, fuse.DirEntry{
			Name: filepath.Base(v.path),
			Ino:  GetInode(v.path),
			Mode: v.getMode(),
		})
	}
	log.WithFields(log.Fields{"direntries": direntries}).Debug("log values")
	return fs.NewListDirStream(direntries), fs.OK
}

//Lookuper
func (sf *FIOTest) Lookup(n *SfsNode, ctx context.Context, name string, out *fuse.EntryOut) (node *fs.Inode, errno syscall.Errno) {
	log.WithFields(log.Fields{
		"n":          n,
		"n.npath":    n.npath,
		"name":       name,
		"out.NodeId": out.NodeId}).Debug("log values")

	// is it the root path?
	fullname := filepath.Join(n.npath, name)
	tn := testnodes.getTestNodeByPath(fullname)
	if tn == nil {
		log.WithFields(log.Fields{
			"n":          n,
			"n.npath":    n.npath,
			"name":       name,
			"out.NodeId": out.NodeId,
			"fullname":   fullname,
			"tn":         testnodes.getTestNodeByPath(fullname)}).Error("could not retrieve testnode tn by fullname")
		return nil, syscall.ENOENT
	}
	log.WithFields(log.Fields{
		"n":          n,
		"n.npath":    n.npath,
		"name":       name,
		"out.NodeId": out.NodeId,
		"tn":         tn,
		"inode":      GetInode(tn.path),
		"mode":       tn.getMode()}).Debug("log values")

	// if true, then get an inode for it
	stable := fs.StableAttr{
		Mode: tn.getMode(),
		Ino:  GetInode(tn.path),
	}
	log.WithFields(log.Fields{
		"n":          n,
		"n.npath":    n.npath,
		"name":       name,
		"out.NodeId": out.NodeId,
		"tn":         tn,
		"inode":      GetInode(tn.path),
		"mode":       tn.getMode(),
		"stable":     stable}).Debug("log values")
	operations := NewNode(filepath.Join(n.npath, name))
	child := n.NewInode(ctx, operations, stable)
	out.NodeId = GetInode(tn.path)
	return child, fs.OK
}

//Opener
func (sf *FIOTest) Open(n *SfsNode, ctx context.Context, flags uint32) (fh fs.FileHandle, fuseFlags uint32, errno syscall.Errno) {
	return nil, 0, 0
}

//Reader
func (sf *FIOTest) Read(n *SfsNode, ctx context.Context, f fs.FileHandle, dest []byte, off int64) (fuse.ReadResult, syscall.Errno) {
	log.WithFields(log.Fields{"n": n, "n.npath": n.npath}).Debug("log values")
	tn := testnodes.getTestNodeByPath(n.npath)
	results := fuse.ReadResultData(tn.content)
	log.WithFields(log.Fields{
		"n":       n,
		"n.npath": n.npath,
		"tn":      tn,
		"results": tn.content}).Debug("log values")
	return results, fs.OK
}

// GetAttrer
var _ = (fs.NodeGetattrer)((*SfsNode)(nil))

func (sf *FIOTest) Getattr(n *SfsNode, ctx context.Context, fh fs.FileHandle, out *fuse.AttrOut) syscall.Errno {
	log.WithFields(log.Fields{"n": n, "n.npath": n.npath}).Debug("log values")

	testnodes.print()
	tn := testnodes.getTestNodeByPath(n.npath)
	log.WithFields(log.Fields{
		"n":       n,
		"n.npath": n.npath,
		"tn":      tn}).Debug("log values")
	if tn == nil {
		log.WithFields(log.Fields{
			"n":       n,
			"n.npath": n.npath,
			"tn":      tn}).Error("could not retrieve testnode tn by n.npath")
		return syscall.ENOENT
	}
	//if !tn.isfile {
	//	return syscall.EISDIR
	//}

	if tn.isfile {
		out.Size = uint64(len(tn.content))
	}
	out.Ino = GetInode(n.npath)
	return fs.OK
}

// FIOPath returns name of implemented FIO Plugin
// sf := Secretsfs Fio
func (sf *FIOTest) FIOPath() string {
	return "tests"
}

func init() {
	fioroot := FIOTest{}
	fm := FIOMap{
		Root: &fioroot,
	}
	RegisterRoot(&fm)
}
