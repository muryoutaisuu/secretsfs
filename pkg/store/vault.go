package store

import (
	"fmt"
	"strconv"
	"io/ioutil"
	"path/filepath"
	"os"
	"os/user"
	"strings"
	//"encoding/json"

	//"gopkg.in/yaml.v2"

	"github.com/hashicorp/vault/api"
	"github.com/hanwen/go-fuse/fuse"
	"github.com/hanwen/go-fuse/fuse/nodefs"
)

const (
	// taken from https://www.vaultproject.io/api/secret/kv/kv-v2.html
	MTDATA = "secret/metadata/"
	DTDATA = "secret/data/"
)

//type authParameter struct {
//	Role_id string `yaml:"role_id"`
//	Secret_id string `yaml:"secret_id"`
//}

type Vault struct {
	client *api.Client
	//TokenAuth *api.Client.Auth().Token()
}

func (v *Vault) GetAttr(name string, context *fuse.Context) (*fuse.Attr, fuse.Status) {
	Log.Debug.Printf("GetAttr name: %v\n",name)
	name = MTDATA + name
	if name == MTDATA { //opening directory (aka secretsfiles/
    return &fuse.Attr{
      Mode: fuse.S_IFDIR | 0550,
    }, fuse.OK
	}
	//if name == "secret/metadata/" {
  //  return &fuse.Attr{
  //    Mode: fuse.S_IFDIR | 0550,
  //  }, fuse.OK
	//}
	if name == MTDATA+"hello" || name == MTDATA+"mury" {
    return &fuse.Attr{
      Mode: fuse.S_IFREG | 0644, Size: uint64(len(name)),
    }, fuse.OK
	}
  Log.Warn.Print(name +" does not exist")
  return nil, fuse.ENOENT
}

func (v *Vault) OpenDir(name string, context *fuse.Context) ([]fuse.DirEntry, fuse.Status) {
	//return []fuse.DirEntry{}, fuse.OK
	Log.Debug.Printf("GetAttr name=\"%v\"\n",name)
	name = MTDATA + name
	s,err := v.client.Logical().List(name)
	if err != nil {
		Log.Error.Print(err)
		return nil, fuse.EIO
	}
	Log.Debug.Printf("GetAttr name=\"%v\" secret=\"%v\" secret.Data=\"%v\"\n",name,s,s.Data)
	dirs := []fuse.DirEntry{}
	// https://github.com/asteris-llc/vaultfs/blob/master/fs/root.go
	// TODO: add Error Handling
	for i := 0; i < len(s.Data["keys"].([]interface{})); i++ {
		d := fuse.DirEntry{
			Name:  s.Data["keys"].([]interface{})[i].(string),
			Mode: fuse.S_IFREG,
		}
		dirs = append(dirs, d)
	}
	return dirs,fuse.OK
}

func (v *Vault) Open(name string, flags uint32, context *fuse.Context) (nodefs.File, fuse.Status) {
	Log.Debug.Printf("Open name=\"%v\"\n",name)
	name = DTDATA + name
	s,err := v.client.Logical().Read(name)
	if err != nil {
		Log.Error.Print(err)
		return nil, fuse.EIO
	}
	Log.Debug.Printf("Open name=\"%v\" secret=\"%v\" secret.Data=\"%v\"\n",name,s,s.Data)
	//for i := 0; i < len(s.Data["keys"].([]interface{})); i++ {
	data := s.Data["data"].([]interface{})
	return nodefs.NewDataFile([]byte(data[0].(string))), fuse.OK




	if name == "secret/hello" {
		err := v.setToken(context)
		if err != nil {
			Log.Error.Print(err)
			return nil, fuse.EIO
		}
		u,err := user.LookupId(strconv.Itoa(int(context.Owner.Uid)))
		if err != nil {
			Log.Error.Print(err)
			return nil, fuse.EIO
		}
		a,err := v.getAccessToken(u)
		if err != nil {
			Log.Error.Printf("msg=\"could not load accessToken\" accessTokenValue=\"%v\"\n",a)
			return nil, fuse.EIO
		}
		//return nodefs.NewDataFile([]byte("mystring")), fuse.OK
		return nodefs.NewDataFile([]byte("mystring")), fuse.OK
	}
  if flags&fuse.O_ANYWRITE != 0 {
    return nil, fuse.EPERM
  }
	return nil,fuse.ENOENT
}

func (v *Vault) String() (string) {
	return "Vault"
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
	v.client.SetToken(a.Auth.ClientToken)
	Log.Debug.Print(v.client.Token())
	return nil
}

func (v *Vault) getAccessToken(u *user.User) (*api.Secret, error) {
	auth,err := v.readAuthToken(u)
	if err != nil {
		Log.Error.Print(err)
		return &api.Secret{}, err
	}
	// https://groups.google.com/forum/#!topic/vault-tool/-4F2RLnGrSE
	postdata := map[string]interface{}{
		"role_id": auth,
	}
	Log.Debug.Printf("login_payload=%v\n",postdata)
	resp,err := v.client.Logical().Write("auth/approle/login", postdata)
	Log.Debug.Printf("resp=%v Data=%v\n ClientToken=\"%v\"",resp,resp.Data,resp.Auth.ClientToken)
	if err != nil {
		Log.Error.Print(err)
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
		Log.Error.Print(err)
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
	Log.Debug.Printf("msg=\"reading authToken\" path=\"%v\"\n",path)
	o,err := ioutil.ReadFile(path)
	if err != nil {
		Log.Error.Print(err)
		return "",err
	}
	authToken := strings.TrimSuffix(string(o), "\n")
	Log.Debug.Printf("msg=\"authToken successfully read\" path=\"%v\"\n",path)
	return authToken,nil
}








func init() {
	c,err := api.NewClient(&api.Config{
		Address: os.Getenv("VAULT_ADDR"),
	})
	if err != nil {
		Log.Error.Fatal(err)
	}
	v := Vault{
		client: c,
	}
	RegisterStore(&v) //https://stackoverflow.com/questions/40823315/x-does-not-implement-y-method-has-a-pointer-receiver
}

