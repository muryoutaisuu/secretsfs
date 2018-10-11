package sfuse

import (
	//"bazil.org/fuse"
	"bazil.org/fuse/fs"
	//"golang.org/x/net/context"
)

type FS struct {}

func (f *FS) Root() (fs.Node, error) {
  n := &Dir{Name: "root"}
  return n, nil
}

var _ fs.FS = (*FS)(nil)

//func (f *FS) Statfs(ctx context.Context, req *fuse.StatfsRequest, resp *fuse.StatfsResponse) error {
//	return nil
//}
