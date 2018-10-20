package store

import (
	"fmt"
	"strconv"
	"io/ioutil"
	"path/filepath"
	"log"
	"os"
	"os/user"
	"strings"
	"encoding/json"

	//"gopkg.in/yaml.v2"

	"github.com/hashicorp/vault/api"
	"github.com/hanwen/go-fuse/fuse"
	"github.com/hanwen/go-fuse/fuse/nodefs"
)

type authParameter struct {
	Role_id string `yaml:"role_id"`
	Secret_id string `yaml:"secret_id"`
}

type Vault struct {
	client *api.Client
	//TokenAuth *api.Client.Auth().Token()
}

func (v *Vault) GetAttr(name string, context *fuse.Context) (*fuse.Attr, fuse.Status) {
	if name == "" {
    return &fuse.Attr{
      Mode: fuse.S_IFDIR | 0550,
    }, fuse.OK
	}
	if name == "test.txt" {
    return &fuse.Attr{
      Mode: fuse.S_IFREG | 0644, Size: uint64(len(name)),
    }, fuse.OK
	}
  log.Print(name +" does not exist")
  return nil, fuse.ENOENT
}

func (v *Vault) OpenDir(name string, context *fuse.Context) ([]fuse.DirEntry, fuse.Status) {
	//return []fuse.DirEntry{}, fuse.OK
	return []fuse.DirEntry{{Name: "test.txt", Mode: fuse.S_IFREG}}, fuse.OK
	//return nil, fuse.ENOENT
}

func (v *Vault) Open(name string, flags uint32, context *fuse.Context) (nodefs.File, fuse.Status) {
	if name == "test.txt" {
		//return nodefs.NewDataFile([]byte(os.Getenv("VAULT_ADDR"))), fuse.OK
		v.setToken(context)
		u,err := user.LookupId(strconv.Itoa(int(context.Owner.Uid)))
		if err != nil {
			log.Print(err)
			return nil, fuse.EIO
		}
		s,err := v.getAccessToken(u)
		if err != nil {
			log.Print("msg=\"could not load accessToken\"")
			log.Print(s)
			return nil, fuse.EIO
		}
		return nodefs.NewDataFile([]byte("mystring")), fuse.OK
		//s,_ := v.client.Auth().Token().LookupSelf()
		//mys,_ := s.TokenID()
		//return nodefs.NewDataFile([]byte(mys)), fuse.OK
	}
  if flags&fuse.O_ANYWRITE != 0 {
    return nil, fuse.EPERM
  }
  return nodefs.NewDataFile([]byte(name)), fuse.OK
}

func (v *Vault) String() (string, error) {
	return "Vault",nil
}


func (v *Vault) setToken(context *fuse.Context) error {
	u,err := user.LookupId(strconv.Itoa(int(context.Owner.Uid)))
	if err != nil {
		return err
	}
	a,err := v.getAccessToken(u)
	if err != nil {
		return err
	}
	jsonData, err := json.Marshal(a.Data)
	if err != nil {
		return err
	}
	v.client.SetToken(string(jsonData))
	return nil
}

func (v *Vault) getAccessToken(u *user.User) (*api.Secret, error) {
	auth,err := v.readAuthToken(u)
	if err != nil {
		log.Print(err)
		return &api.Secret{}, err
	}
	// https://groups.google.com/forum/#!topic/vault-tool/-4F2RLnGrSE
	data := map[string]interface{}{
		"role_id": auth,
	}
	resp,err := v.client.Logical().Write("auth/approle/login", data)
	if err != nil {
		log.Print(err)
		return &api.Secret{}, err
	}
	if resp.Auth == nil {
		return resp, fmt.Errorf("no auth info returned")
	}
	return resp,err
}

func (v *Vault) secret(u *user.User) (*api.Secret, error) {
	authToken,err := v.readAuthToken(u)
	if err != nil {
		log.Print(err)
		return &api.Secret{}, err
	}
	c := v.client
	auth := c.Auth()
	tokenauth := auth.Token()
	secret,err := tokenauth.Lookup(authToken)
	return secret,err
}

func (v *Vault) readAuthToken(u *user.User) (string, error) {
	path := filepath.Join(u.HomeDir, os.Getenv("SECRETSFS_FILE_ROLEID"))
	o,err := ioutil.ReadFile(path)
	if err != nil {
		log.Print(err)
		return "",err
	}
	authToken := strings.TrimSuffix(string(o), "\n")
	return authToken,nil
}





func init() {
	c,err := api.NewClient(&api.Config{
		Address: os.Getenv("VAULT_ADDR"),
	})
	if err != nil {
		log.Fatal(err)
	}
	v := Vault{
		client: c,
	}
	RegisterStore(&v) //https://stackoverflow.com/questions/40823315/x-does-not-implement-y-method-has-a-pointer-receiver
}

