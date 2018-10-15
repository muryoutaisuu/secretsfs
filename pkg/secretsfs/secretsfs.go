package secretsfs

// after the example: https://github.com/hanwen/go-fuse/blob/master/zipfs/memtree.go

import (
	"fmt"
	"strings"


	"github.com/hanwen/go-fuse/fuse"
	"github.com/hanwen/go-fuse/fuse/nodefs"
	"github.com/Muryoutaisuu/secretsfs/pkg/fio"
)

type SFile interface {
	Stat(out *fuse.Attr)
	Data() []byte
}

type Secret struct {
	SFile
}

type sNode struct {
	nodefs.Node
	file SFile
	fs *Secretsfs
}

type Secretsfs struct {
	root *sNode
	files map[string]SFile
	Name string
}

func NewSecretsfs(fms map[string]*fio.FIOMap) *Secretsfs {
	fs := &Secretsfs{
		root: &sNode{Node : nodefs.NewDefaultNode()},
		Name: "root",
	}
	fs.root.fs = fs
	return fs
}

func (fs *Secretsfs) String() string {
	return fs.Name
}

func (fs *Secretsfs) Root() nodefs.Node {
	return fs.root
}

func (fs *Secretsfs) onMount() {
	for k, v := range fs.files {
		fs.addFile(k, v)
	}
	fs.files = nil
}

func (n *sNode) OnMount(c *nodefs.FileSystemConnector) {
	n.fs.onMount()
}

func (n *sNode) Print(indent int) {
	s := ""
	for i := 0; i < indent; i++ {
		s = s + " "
	}

	children := n.Inode().Children()
	for k, v := range children {
		if v.IsDir() {
			fmt.Println(s + k + ":")
			mn, ok := v.Node().(*sNode)
			if ok {
				mn.Print(indent + 2)
			}
		} else {
			fmt.Println(s + k)
		}
	}
}

func (n *sNode) OpenDir(context *fuse.Context) (stream []fuse.DirEntry, code fuse.Status) {
	children := n.Inode().Children()
	stream = make([]fuse.DirEntry, 0, len(children))
	for k, v := range children {
		mode := fuse.S_IFREG | 0666
		if v.IsDir() {
			mode = fuse.S_IFDIR | 0777
		}
		stream = append(stream, fuse.DirEntry{
			Name: k,
			Mode: uint32(mode),
		})
	}
	return stream, fuse.OK
}

func (n *sNode) Open(flags uint32, context *fuse.Context) (fuseFile nodefs.File, code fuse.Status) {
	if flags&fuse.O_ANYWRITE != 0 {
		return nil, fuse.EPERM
	}
	return nodefs.NewDataFile(n.file.Data()), fuse.OK
}

func (n *sNode) Deletable() bool {
	return false
}

func (n *sNode) GetAttr(out *fuse.Attr, file nodefs.File, context *fuse.Context) fuse.Status {
	if n.Inode().IsDir() {
		out.Mode = fuse.S_IFDIR | 0777
		return fuse.OK
	}
	n.file.Stat(out)
	out.Blocks = (out.Size + 511) / 512
	return fuse.OK
}

func (n *Secretsfs) addRootChildren(fms map[string]*fio.FIOMap) error {
	for k := range fms {
		n.addFile(k,Secret{})
	}
	return nil
}

func (n *Secretsfs) addFile(name string, f SFile) error {
	comps := strings.Split(name, "/")

	node := n.root.Inode()
	for i, c := range comps {
		child := node.GetChild(c)
		if child == nil {
			fsnode := &sNode{
				Node: nodefs.NewDefaultNode(),
				fs:   n,
			}
			if i == len(comps)-1 {
				fsnode.file = f
			}

			child = node.NewChild(c, fsnode.file == nil, fsnode)
		}
		node = child
	}
	return nil
}
