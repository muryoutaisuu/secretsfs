package secretsfs

// taken from example https://github.com/bazil/fuse/blob/master/examples/clockfs/clockfs.go


import (
	"fmt"
	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	"github.com/Muryoutaisuu/secretsfs/pkg/secretsfs/sfuse"
)

func Run(mountpoint string) error {
	c, err := fuse.Mount(
		mountpoint,
		fuse.FSName("clock"),
		fuse.Subtype("clockfsfs"),
		fuse.LocalVolume(),
		fuse.VolumeName("Clock filesystem"),
	)
	if err != nil {
		return err
	}
	defer c.Close()

	if p := c.Protocol(); !p.HasInvalidate() {
		return fmt.Errorf("kernel FUSE support is too old to have invalidations: version %v", p)
	}

	srv := fs.New(c, nil)
	filesys := &sfuse.FS{}

	if err := srv.Serve(filesys); err != nil {
    return err
  }

	<-c.Ready
  if err := c.MountError; err != nil {
    return err
  }

	return nil
}
