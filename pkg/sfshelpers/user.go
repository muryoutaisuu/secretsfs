package sfshelpers

import (
	"github.com/hanwen/go-fuse/fuse"
	"strconv"
	"os/user"
)

// getUser returns user element
// used for getting userinfo for logging
func GetUser(context *fuse.Context) (*user.User, error) {
        return user.LookupId(strconv.Itoa(int(context.Owner.Uid)))
}
