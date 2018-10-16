package store

import (
	"io/ioutil"
	"path/filepath"
	"log"
	"os"
	"os/user"
	"strings"

	"github.com/hashicorp/vault/api"
	"github.com/hanwen/go-fuse/fuse"
	"github.com/hanwen/go-fuse/fuse/nodefs"
)


type Vault struct {
	client *api.Client
}

func (v *Vault) GetAttr(name string, context *fuse.Context) (*fuse.Attr, fuse.Status) {
  switch name {
  case "test.txt":
    return &fuse.Attr{
      Mode: fuse.S_IFREG | 0644, Size: uint64(len(name)),
    }, fuse.OK
  case "":
    return &fuse.Attr{
      Mode: fuse.S_IFDIR | 0755,
    }, fuse.OK
  }
  log.Fatal(name +" does not exist")
  return nil, fuse.ENOENT
}

func (v *Vault) OpenDir(name string, context *fuse.Context) ([]fuse.DirEntry, fuse.Status) {
	//return []fuse.DirEntry{}, fuse.OK
	return []fuse.DirEntry{{Name: "test.txt", Mode: fuse.S_IFREG}}, fuse.OK
	//return nil, fuse.ENOENT
}

func (v *Vault) Open(name string, flags uint32, context *fuse.Context) (nodefs.File, fuse.Status) {
	if name == "test.txt" {
		return nodefs.NewDataFile([]byte(os.Getenv("VAULT_ADDR"))), fuse.OK
	}
  if flags&fuse.O_ANYWRITE != 0 {
    return nil, fuse.EPERM
  }
  return nodefs.NewDataFile([]byte(name)), fuse.OK
}

func (v *Vault) String() (string, error) {
	return "Vault",nil
}





func (v *Vault) secret(u *user.User) (*api.Secret, error) {
	authToken,_ := readAuthToken(u)
	c := v.client
	auth := c.Auth()
	tokenauth := auth.Token()
	secret,err := tokenauth.Lookup(authToken)
	return secret,err
}

func readAuthToken(u *user.User) (string, error) {
	path := filepath.Join(u.HomeDir, "/authTokenfile")
	o,err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}
	authToken := strings.TrimSuffix(string(o), "\n")
	return authToken,nil
}






func init() {
	c,_ := api.NewClient(&api.Config{
		Address: os.Getenv("VAULT_ADDR"),
	})
	v := Vault{
		client: c,
	}
	RegisterStore(&v) //https://stackoverflow.com/questions/40823315/x-does-not-implement-y-method-has-a-pointer-receiver
}

