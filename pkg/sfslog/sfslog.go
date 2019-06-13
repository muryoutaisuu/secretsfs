// sfslog package provides all needed logging features for a consistent logging
// format through all provided plugins.
// Code taken from example:
// https://www.ardanlabs.com/blog/2013/11/using-log-package-in-go.html
package sfslog

import (
	//"os"
	"os/user"
	//"io"
	//"io/ioutil"
	// TODO: make debug configurable

	log "github.com/sirupsen/logrus"
)

func DefaultEntry(name string, user *user.User) *log.Entry {
	return log.WithFields(log.Fields{
		"name":     name,
		"userid":   user.Uid,
		"username": user.Username,
	})
}
