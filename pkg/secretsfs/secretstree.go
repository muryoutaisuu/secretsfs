package secretsfs

// after the example: https://github.com/hanwen/go-fuse/blob/master/zipfs/memtree.go

import (
	"fmt"
	"strings"


	"github.com/hanwen/go-fuse/fuse"
	"github.com/hanwen/go-fuse/fuse/nodefs"
)

type SFile interface {
	Stat(out *fuse.Attr)
	Data() []byte
}

type sNode struct {
	nodefs.Node
	file SFile
	fs *SecretsTreeFS
}

type SecretsTreeFS struct {
	root *sNode
	files map[string]SFile
	Name string
}

func NewSecretsTreeFS(files map[string]SFile) *SecretsTreeFS {
	fs := &SecretsTreeFS{
		root: &sNode{Node : nodefs.NewDefaultNode()},
		files: files,
		Name: "root",
	}
	fs.root.fs = fs
	return fs
}

func (fs *SecretsTreeFS) String() string {
	return fs.Name
}

func (fs *SecretsTreeFS) Root() nodefs.Node {
	return fs.root
}

func (fs *SecretsTreeFS) onMount() {
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

func (n *SecretsTreeFS) addFile(name string, f SFile) {
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
}
