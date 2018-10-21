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
	if name == "" {
    return &fuse.Attr{
      Mode: fuse.S_IFDIR | 0550,
    }, fuse.OK
	}
	if name == "secret" {
    return &fuse.Attr{
      Mode: fuse.S_IFDIR | 0550,
    }, fuse.OK
	}
	if name == "secret/hello" {
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
	name = "secret/metadata/" + name
	s,err := v.client.Logical().List(name)
	if err != nil {
		Log.Error.Print(err)
		return nil, fuse.EIO
	}
	Log.Debug.Printf("GetAttr name=\"%v\" secret=\"%v\" secret.Data=\"%v\"\n",name,s,s.Data)

	data := s.Data
	Log.Debug.Printf("data=\"%v\"\n",data["keys"])
	keys := data["keys"]
	Log.Info.Printf("keys=\"%v\" keysType=\"%T\" keys0Type=\"%T\"\n",keys,keys)
	d,ok := keys.([]api.Secret)
	if ok != true {
		Log.Error.Printf("dataconversion ok=\"%v\"\n",ok)
		Log.Debug.Printf("dataconversion d=\"%v\"\n",d)
		test := []string{"myfirst", "mysecond"}
		Log.Debug.Printf("dataconversion test=\"%v\"\n",test)
		return nil, fuse.EIO
	}

	Log.Debug.Printf("d=\"%v\" dType=\"%T\"\n",d,d)

	//k := s.Data["keys"]
	//Log.Debug.Printf("k=\"%v\"\n",k)
	//Log.Debug.Printf("k=\"%v\"\n",k[0])
	if name == "" {
		return []fuse.DirEntry{{Name: "secret", Mode: fuse.S_IFDIR}}, fuse.OK
	}
	if name == "secret" {
		return []fuse.DirEntry{{Name: "hello", Mode: fuse.S_IFREG}}, fuse.OK
	}
	return nil, fuse.ENOENT
}

func (v *Vault) Open(name string, flags uint32, context *fuse.Context) (nodefs.File, fuse.Status) {
	Log.Debug.Printf("GetAttr name: %v\n",name)
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
  return nodefs.NewDataFile([]byte(name)), fuse.OK
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

