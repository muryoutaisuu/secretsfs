package fusehelpers

import (
	"context"
	"os/user"
	"strconv"

	"github.com/hanwen/go-fuse/v2/fuse"
)

// GetUserFromContext returns the user that called the filesystem operation
func GetUserFromContext(ctx context.Context) (*user.User, error) {
	c := ctx.(*fuse.Context)
	u, err := user.LookupId(strconv.Itoa(int(c.Caller.Owner.Uid)))
	return u, err
}
