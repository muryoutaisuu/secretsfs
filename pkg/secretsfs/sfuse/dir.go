package sfuse

import (
  "bazil.org/fuse"
	"golang.org/x/net/context"
  "bazil.org/fuse/fs"
)

type Dir struct {
	Name string
}

func (d *Dir) Attr(ctx context.Context, attr *fuse.Attr) error {
	attr.Mode = 0550
	return nil
}

var _ fs.Node = (*Dir)(nil)

