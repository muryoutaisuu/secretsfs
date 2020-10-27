package secretsfs

import (
	"context"
	"path/filepath"
	"strconv"
	"syscall"

	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
	log "github.com/sirupsen/logrus"

	"github.com/muryoutaisuu/secretsfs/pkg/store"
)

type FIOSecretsFiles struct{}

var _ = (FIORoot)((*FIOSecretsFiles)(nil))

func (sf *FIOSecretsFiles) Readdir(n *SfsNode, ctx context.Context) (out fs.DirStream, errno syscall.Errno) {
	log.WithFields(log.Fields{"n": n, "n.npath": n.npath}).Debug("log values")

	sto := *store.GetStore()
	_, secpath := rootName(n.npath)
	sec, err := sto.GetSecret(secpath, ctx)
	if err != nil {
		log.WithFields(log.Fields{"secpath": secpath, "error": err, "calling": "sto.GetSecret(secpath, ctx)"}).Error("Got error while getting secret")
		return nil, syscall.ENOENT
	}
	if sec.Mode != fuse.S_IFDIR {
		log.WithFields(log.Fields{"secpath": secpath, "secret": sec, "sec.Mode": strconv.FormatInt(int64(sec.Mode), 16)}).Debug("secret is not a directory type")
		return nil, syscall.ENOTDIR
	}

	var direntries []fuse.DirEntry

	log.Println("logging subs")
	for _, v := range sec.Subs {
		fixedpath := sf.prefixPath(v.Path)
		log.WithFields(log.Fields{
			"v.Path":    v.Path,
			"v.Mode":    strconv.FormatInt(int64(v.Mode), 16),
			"fixedpath": fixedpath,
			"Name":      filepath.Base(fixedpath),
			"Ino":       GetInode(fixedpath)}).Trace("logging subs")
		direntries = append(direntries, fuse.DirEntry{
			Name: filepath.Base(fixedpath),
			Ino:  GetInode(fixedpath),
			Mode: uint32(v.Mode),
		})
	}
	log.WithFields(log.Fields{"direntries": direntries}).Debug("log values")
	return fs.NewListDirStream(direntries), fs.OK
}

func (sf *FIOSecretsFiles) Lookup(n *SfsNode, ctx context.Context, name string, out *fuse.EntryOut) (node *fs.Inode, errno syscall.Errno) {
	log.WithFields(log.Fields{
		"n":          n,
		"n.npath":    n.npath,
		"name":       name,
		"out.NodeId": out.NodeId}).Debug("log values")

	// is it the root path?
	sto := *store.GetStore()
	_, secpath := rootName(n.npath)
	fullname := filepath.Join(secpath, name)
	sec, err := sto.GetSecret(fullname, ctx)
	if err != nil {
		log.WithFields(log.Fields{
			"calling":  "sto.GetSecret(fullname, ctx)",
			"fullname": fullname,
			"n":        n,
			"n.npath":  n.npath,
			"name":     name,
			"error":    err}).Warn("got error while getting secret, probably not enough permissions")
		//return nil, syscall.EPERM
		sec = &store.Secret{Path: fullname, Mode: fuse.S_IFDIR, Content: "", Subs: nil}
	}
	prefixedfullname := sf.prefixPath(fullname)
	log.WithFields(log.Fields{"inode": GetInode(prefixedfullname), "mode": strconv.FormatInt(int64(sec.Mode), 16)}).Debug("log values")

	// if true, then get an inode for it
	stable := fs.StableAttr{
		Mode: uint32(sec.Mode),
		Ino:  GetInode(prefixedfullname),
	}
	log.WithFields(log.Fields{"stable": stable}).Debug("log values")
	operations := NewNode(prefixedfullname)
	child := n.NewInode(ctx, operations, stable)
	out.NodeId = GetInode(prefixedfullname)
	return child, fs.OK
}

func (sf *FIOSecretsFiles) Open(n *SfsNode, ctx context.Context, flags uint32) (fh fs.FileHandle, fuseFlags uint32, errno syscall.Errno) {
	return nil, 0, 0
}

func (sf *FIOSecretsFiles) Read(n *SfsNode, ctx context.Context, f fs.FileHandle, dest []byte, off int64) (fuse.ReadResult, syscall.Errno) {
	log.WithFields(log.Fields{"n": n, "n.npath": n.npath}).Debug("log values")

	sto := *store.GetStore()
	_, secpath := rootName(n.npath)
	sec, err := sto.GetSecret(secpath, ctx)
	if err != nil {
		log.WithFields(log.Fields{"calling": "sto.GetSecret(secpath, ctx)", "secpath": secpath, "error": err}).Error("got error while getting secret")
		return nil, syscall.ENOENT
	}
	results := fuse.ReadResultData([]byte(sec.Content))
	log.WithFields(log.Fields{"results": results}).Debug("log values")
	return results, fs.OK
}

func (sf *FIOSecretsFiles) Getattr(n *SfsNode, ctx context.Context, fh fs.FileHandle, out *fuse.AttrOut) syscall.Errno {
	log.WithFields(log.Fields{
		"n":                   n,
		"n.npath":             n.npath,
		"IsRootPath(n.npath)": IsRootPath(n.npath)}).Debug("log values")

	// if rootpath, then no store is needed
	if IsRootPath(n.npath) {
		out.Ino = GetInode(n.npath)
		return fs.OK
	}

	sto := *store.GetStore()
	_, secpath := rootName(n.npath)
	sec, err := sto.GetSecret(secpath, ctx)
	if err != nil {
		log.WithFields(log.Fields{
			"calling": "sto.GetSecret(secpath, ctx)",
			"secpath": secpath,
			"n":       n,
			"n.npath": n.npath,
			"error":   err}).Warn("got error while getting secret, probably not enough permissions")
		sec = &store.Secret{Path: secpath, Mode: fuse.S_IFDIR + 0x0700, Content: "", Subs: nil}
		//return syscall.ENOENT
	}
	log.WithFields(log.Fields{"inode": GetInode(n.npath), "Mode": strconv.FormatInt(int64(sec.Mode), 16)}).Debug("log values")

	if sec.Mode == fuse.S_IFREG {
		out.Size = uint64(len(sec.Content))
	}
	out.Ino = GetInode(n.npath)
	return fs.OK
}

func (sf *FIOSecretsFiles) FIOPath() string {
	return "secretsfiles"
}

func (sf *FIOSecretsFiles) prefixPath(npath string) string {
	return string(filepath.Separator) + filepath.Join(sf.FIOPath(), npath)
}

func init() {
	fioroot := FIOSecretsFiles{}
	fm := FIOMap{
		Root: &fioroot,
	}
	RegisterRoot(&fm)
}
