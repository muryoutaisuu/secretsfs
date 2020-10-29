package secretsfs

import (
	"context"
	"fmt"
	"os/user"
	"path/filepath"
	"syscall"

	fh "github.com/muryoutaisuu/secretsfs/pkg/fusehelpers"
	"github.com/muryoutaisuu/secretsfs/pkg/store"

	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
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

type internalNode struct {
	path          string
	isfile        bool
	needPrivilege bool
	filemode      uint32
	getContent    func(context.Context) []byte
}
type internalNodes struct {
	nodes []*internalNode
}

var internalnodes = internalNodes{
	[]*internalNode{
		&internalNode{"/internal", false, false, 0755, nil},
		&internalNode{"/internal/inodes", true, true, 0750, prettyprintInodes},
		&internalNode{"/internal/user", true, false, 0755, prettyprintUser},
		&internalNode{"/internal/privileged", true, false, 0755, prettyprintIsPrivileged},
		&internalNode{"/internal/store", false, false, 0755, nil},
		&internalNode{"/internal/store/vault_kv", true, true, 0750, prettyprintVault},
		&internalNode{"/internal/store/useroverrides", true, true, 0750, prettyprintUseroverrides},
		&internalNode{"/internal/store/useroverride", true, false, 0755, prettyprintUseroverride},
	},
}

func prettyprintInodes(ctx context.Context) []byte {
	content, err := PrettyPrint(pathsInodes)
	if err != nil {
		return []byte(fmt.Sprintf("got error on prettyprinting, err=\"%v\"\n", err))
	}
	return content
}

func prettyprintUser(ctx context.Context) []byte {
	u, err := fh.GetUserFromContext(ctx)
	content, err := PrettyPrint(u)
	if err != nil {
		return []byte(fmt.Sprintf("got error on prettyprinting, err=\"%v\"\n", err))
	}
	return content
}

func prettyprintIsPrivileged(ctx context.Context) []byte {
	return []byte(fmt.Sprintf("%v\n", isPrivileged(ctx)))
}

func prettyprintVault(ctx context.Context) []byte {
	if s := *store.GetStore(); s.String() != "vault_kv" {
		return []byte(fmt.Sprintf("vault is not the configured store, currently configured store: \"%v\"\n", s.String()))
	}
	pfvc, err := store.GetClient(ctx)
	if err != nil {
		return []byte(fmt.Sprintf("got error while calling store.GetClient(ctx), err=\"%v\"\n", err))
	}
	content, err := PrettyPrint(pfvc)
	if err != nil {
		return []byte(fmt.Sprintf("got error on prettyprinting, err=\"%v\"\n", err))
	}
	return content
}

func prettyprintUseroverrides(ctx context.Context) []byte {
	return []byte(fmt.Sprintf("%v\n", viper.GetStringMapString("store.vault.roleid.useroverride")))
}
func prettyprintUseroverride(ctx context.Context) []byte {
	u, _ := fh.GetUserFromContext(ctx)
	finalizedpath := store.FinIdPath(u)
	ufinpath := struct {
		username string
		finpath  string
	}{
		u.Name,
		finalizedpath,
	}
	return []byte(fmt.Sprintf("%v\n", ufinpath))
}

func (t *internalNodes) isDir(npath string) bool {
	for _, v := range t.nodes {
		if !v.isfile && v.path == npath {
			return true
		}
	}
	return false
}
func (t *internalNodes) isFile(npath string) bool {
	for _, v := range t.nodes {
		if v.isfile && v.path == npath {
			return true
		}
	}
	return false
}
func (t *internalNodes) getInternalNodeByPath(npath string) *internalNode {
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
func (t *internalNodes) getDirEntries(npath string) (direntries []*internalNode) {
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
func (t *internalNode) getMode() uint32 {
	if t.isfile {
		return fuse.S_IFREG
	}
	return fuse.S_IFDIR
}

type FIOInternal struct{}

var _ = (FIORoot)((*FIOInternal)(nil))

//Readdirer
func (sf *FIOInternal) Readdir(n *SfsNode, ctx context.Context) (out fs.DirStream, errno syscall.Errno) {
	log.WithFields(log.Fields{"n": n, "n.npath": n.npath}).Debug("log values")
	in := internalnodes.getInternalNodeByPath(n.npath)
	if in.needPrivilege && !isPrivileged(ctx) {
		u, _ := fh.GetUserFromContext(ctx)
		log.WithFields(log.Fields{"n": n, "n.npath": n.npath, "username": u.Name}).Error("User is not privileged")
		return nil, syscall.EPERM
	}
	if !internalnodes.isDir(n.npath) {
		log.WithFields(log.Fields{"n.npath": n.npath}).Error("node is not a directory")
		return nil, syscall.ENOENT
	}
	log.WithFields(log.Fields{"n.npath": n.npath, "condition": "internalnodes.isDir(n.npath) == true"}).Debug("node is a directory")

	entries := internalnodes.getDirEntries(n.npath)
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
func (sf *FIOInternal) Lookup(n *SfsNode, ctx context.Context, name string, out *fuse.EntryOut) (node *fs.Inode, errno syscall.Errno) {
	log.WithFields(log.Fields{
		"n":          n,
		"n.npath":    n.npath,
		"name":       name,
		"out.NodeId": out.NodeId}).Debug("log values")

	// is it the root path?
	fullname := filepath.Join(n.npath, name)
	in := internalnodes.getInternalNodeByPath(fullname)
	if in == nil {
		log.WithFields(log.Fields{
			"n":          n,
			"n.npath":    n.npath,
			"name":       name,
			"out.NodeId": out.NodeId,
			"fullname":   fullname,
			"in":         internalnodes.getInternalNodeByPath(fullname)}).Error("could not retrieve internalnode in by fullname")
		return nil, syscall.ENOENT
	}
	log.WithFields(log.Fields{
		"n":          n,
		"n.npath":    n.npath,
		"name":       name,
		"out.NodeId": out.NodeId,
		"in":         in,
		"inode":      GetInode(in.path),
		"mode":       in.getMode()}).Debug("log values")

	// if true, then get an inode for it
	stable := fs.StableAttr{
		Mode: in.getMode(),
		Ino:  GetInode(in.path),
	}
	log.WithFields(log.Fields{
		"n":          n,
		"n.npath":    n.npath,
		"name":       name,
		"out.NodeId": out.NodeId,
		"in":         in,
		"inode":      GetInode(in.path),
		"mode":       in.getMode(),
		"stable":     stable}).Debug("log values")
	operations := NewNode(filepath.Join(n.npath, name))
	child := n.NewInode(ctx, operations, stable)
	out.NodeId = GetInode(in.path)
	return child, fs.OK
}

//Opener
func (sf *FIOInternal) Open(n *SfsNode, ctx context.Context, flags uint32) (fh fs.FileHandle, fuseFlags uint32, errno syscall.Errno) {
	in := internalnodes.getInternalNodeByPath(n.npath)
	if in.needPrivilege && !isPrivileged(ctx) {
		return nil, 0, syscall.EPERM
	}
	return nil, 0, 0
}

//Reader
func (sf *FIOInternal) Read(n *SfsNode, ctx context.Context, f fs.FileHandle, dest []byte, off int64) (fuse.ReadResult, syscall.Errno) {
	log.WithFields(log.Fields{"n": n, "n.npath": n.npath}).Debug("log values")
	in := internalnodes.getInternalNodeByPath(n.npath)
	if in.needPrivilege && !isPrivileged(ctx) {
		return nil, syscall.EPERM
	}
	content := in.getContent(ctx)
	results := fuse.ReadResultData(content)
	log.WithFields(log.Fields{
		"n":       n,
		"n.npath": n.npath,
		"in":      in,
		"results": string(content)}).Debug("log values")
	return results, fs.OK
}

// GetAttrer
var _ = (fs.NodeGetattrer)((*SfsNode)(nil))

func (sf *FIOInternal) Getattr(n *SfsNode, ctx context.Context, fh fs.FileHandle, out *fuse.AttrOut) syscall.Errno {
	log.WithFields(log.Fields{"n": n, "n.npath": n.npath}).Debug("log values")

	in := internalnodes.getInternalNodeByPath(n.npath)
	//if in.needPrivilege && !isPrivileged(ctx) {
	//	return syscall.EPERM
	//}

	log.WithFields(log.Fields{
		"n":       n,
		"n.npath": n.npath,
		"in":      in}).Debug("log values")
	if in == nil {
		log.WithFields(log.Fields{
			"n":       n,
			"n.npath": n.npath,
			"in":      in}).Error("could not retrieve internalnode in by n.npath")
		return syscall.ENOENT
	}
	//if !in.isfile {
	//	return syscall.EISDIR
	//}

	if in.isfile {
		out.Size = uint64(len(in.getContent(ctx)))
	}
	out.Mode = in.filemode
	out.Ino = GetInode(n.npath)
	return fs.OK
}

// FIOPath returns name of implemented FIO Plugin
// sf := Secretsfs Fio
func (sf *FIOInternal) FIOPath() string {
	return "internal"
}

func isPrivileged(ctx context.Context) bool {
	u, err := fh.GetUserFromContext(ctx)
	if err != nil {
		log.WithFields(log.Fields{
			"ctx":     ctx,
			"err":     err,
			"calling": "fh.GetUserFromContext(ctx)"}).Error("Got error while getting user from context")
		return false
	}

	// check if user is privileged themselves
	users := viper.GetStringSlice("fio.internal.privileges.users")
	log.WithFields(log.Fields{"users": users}).Debug("log values")
	if users != nil {
		for _, pu := range users {
			if pu == u.Name {
				return true
			}
		}
	}

	// check if user is in privileged group
	pgroups := getGroupsFromNames(viper.GetStringSlice("fio.internal.privileges.groups"))
	usergroupids, err := u.GroupIds()
	if err != nil {
		log.WithFields(log.Fields{
			"u":            u,
			"pgroups":      pgroups,
			"usergroupids": usergroupids,
			"err":          err,
			"calling":      "u.GroupIds()"}).Error("Got error while getting all usergroupids from user")
	}
	usergroupids = append([]string{u.Gid}, usergroupids...)
	for _, ug := range usergroupids {
		if _, ok := pgroups[ug]; ok {
			return true
		}
	}

	// if no match was found, then user is not privileged
	log.WithFields(log.Fields{
		"u":            u,
		"pgroups":      pgroups,
		"usergroupids": usergroupids}).Info("user is not privileged")
	return false
}

func getGroupsFromNames(names []string) map[string]*user.Group {
	var groups = map[string]*user.Group{}
	for _, n := range names {
		g, err := user.LookupGroup(n)
		if err != nil {
			log.WithFields(log.Fields{"names": names, "n": n, "err": err, "calling": "user.LookupGroup(n)"}).Error("Error while looking up group n")
			continue
		}
		groups[g.Gid] = g
	}
	return groups
}

//func getGroupsFromIds(ids []string) (groups map[string]*user.Group) {
//	for _,i := range ids {
//		g, err := user.LookupGroupId(i)
//		if err != nil {
//			log.WithFields(log.Fields{"ids": ids, "i": i, "err": err, "calling": "user.LookupGroupId(i)"}).Error("Error while looking up group i")
//			continue
//		}
//		groups[g.Gid] = g
//	}
//	return
//}

func init() {
	fioroot := FIOInternal{}
	fm := FIOMap{
		Root: &fioroot,
	}
	RegisterRoot(&fm)
}
