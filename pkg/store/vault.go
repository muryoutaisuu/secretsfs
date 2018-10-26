package store

import (
	"errors"
	"fmt"
	"strconv"
	"io/ioutil"
	"path/filepath"
	"os"
	"os/user"
	"strings"
	"path"
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

// Filetype define the type of the returned value element of vault
type Filetype int
const (
	CTrueDir   Filetype = 0 // exists in Vault as a directory
	CValue     Filetype = 1 // Value of a key=value pair
	CFile      Filetype = 2 // Key of a key=value pair, emulated as a directory
	CNull      Filetype = 3 // not a valid vault element
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
	Log.Debug.Printf("ops=GetAttr name=\"%v\"\n",name)
	//name = MTDATA + name

	// opening directory (aka secretsfiles/)
	if name == "" {
    return &fuse.Attr{
      Mode: fuse.S_IFDIR | 0550,
    }, fuse.OK
	}

	// get type
	Log.Debug.Printf("name=\"%v\"\n",name)
	t,err := v.getType(name)
	Log.Debug.Printf("op=GetAttr t=\"%v\" err=\"%v\"\n",t,err)
	if err != nil {
		Log.Error.Printf("op=GetAttr err=\"%v\"\n",err)
		return nil, fuse.EIO
	}

	// act according to type
	switch t {
	case CTrueDir:
		return &fuse.Attr{
      Mode: fuse.S_IFDIR | 0550,
    }, fuse.OK
	case CFile:
		Log.Debug.Printf("op=GetAttr t=CFile\n")
		return &fuse.Attr{
      Mode: fuse.S_IFDIR | 0550,
    }, fuse.OK
	case CValue:
		return &fuse.Attr{
      Mode: fuse.S_IFREG | 0550,
			Size: uint64(len(name)),
    }, fuse.OK
	default:
		return nil, fuse.ENOENT
	}
}

func (v *Vault) OpenDir(name string, context *fuse.Context) ([]fuse.DirEntry, fuse.Status) {
	Log.Debug.Printf("GetAttr name=\"%v\"\n",name)
	//name = MTDATA + name
	t,err := v.getType(name)
	Log.Debug.Printf("ops=OpenDir t=\"%v\" err=\"%v\"\n",t,err)
	if err != nil {
		Log.Error.Print(err)
		return nil, fuse.EIO
	}

	switch t {
	case CTrueDir:
		dirs,err := v.listDir(name)
		if err != nil {
			Log.Error.Print(err)
			return *dirs, fuse.EIO
		}
		Log.Debug.Printf("op=OpenDir name=\"%v%v\" dirs=\"%v\" err=\"%v\"\n",MTDATA,name,dirs,err)
		return *dirs, fuse.OK
	case CFile:
		dirs,err := v.listFile(name)
		Log.Debug.Printf("op=OpenDir dirs=\"%v\" err=\"%v\"\n",dirs,err)
		if err != nil {
			Log.Error.Print(err)
			return nil, fuse.EIO
		}
		Log.Debug.Printf("op=OpenDir ctype=CFile secretType=\"%T\" secret=\"%v\"\n",dirs,dirs)
		return *dirs, fuse.OK
	case CValue:
		return nil, fuse.ENOTDIR
	}
	return nil, fuse.ENOENT
}

func (v *Vault) Open(name string, flags uint32, context *fuse.Context) (nodefs.File, fuse.Status) {
	Log.Debug.Printf("Open name=\"%v\"\n",name)
	//name = DTDATA + name
	s,err := v.client.Logical().Read(DTDATA + name)
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
	Log.Debug.Printf("resp=%v Data=%v\n ClientToken=\"%v\"\n",resp,resp.Data,resp.Auth.ClientToken)
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

func (v *Vault) listDir(name string) (*[]fuse.DirEntry, error) {
	Log.Debug.Printf("op=listDir MTDATA=\"%v\" name=\"%v\"",MTDATA,name)
	s,err := v.client.Logical().List(MTDATA + name)
	Log.Debug.Printf("secret=\"%v\"\n",s)

	// can't list in vault
	if err != nil || s == nil {
		if err == nil {
			err = errors.New("cant list")
		}
		Log.Debug.Print(err)
		return nil, err
	}

	Log.Debug.Printf("GetAttr name=\"%v\" secret=\"%v\" secret.Data=\"%v\"\n",name,s,s.Data)
	dirs := []fuse.DirEntry{}
	// https://github.com/asteris-llc/vaultfs/blob/master/fs/root.go
	// TODO: add Error Handling
	Log.Debug.Printf("op=listDir dirs=\"%v\"\n",dirs)
	for i := 0; i < len(s.Data["keys"].([]interface{})); i++ {
		d := fuse.DirEntry{
			Name:  path.Base(s.Data["keys"].([]interface{})[i].(string)),
			Mode: fuse.S_IFREG,
		}
		dirs = append(dirs, d)
		Log.Debug.Printf("op=listDir dirs=\"%v\"\n",dirs)
	}
	return &dirs,nil
}

func (v *Vault) listFile(name string) (*[]fuse.DirEntry, error) {
	s,err := v.client.Logical().Read(DTDATA + name)
	if err != nil || s == nil {
		if err == nil {
			errors.New("cant read")
		}
		return nil,err
	}
	Log.Debug.Printf("op=listFile secret=\"%v\"\n",s)
	Log.Debug.Printf("op=listFile secret.Data=\"%v\" secret.DataType=\"%T\"\n",s.Data,s.Data)
	data := s.Data["data"].(map[string]interface{})
	Log.Debug.Printf("op=listFile data=\"%v\" dataType=\"%T\"\n",data,data)
	dirs := []fuse.DirEntry{}
	for k := range data {
		d := fuse.DirEntry{
			Name:  data[k].(string),
			Mode: fuse.S_IFREG,
		}
		dirs = append(dirs, d)
		Log.Debug.Printf("op=listFile dirs=\"%v\"\n",dirs)
	}
	return &dirs,nil
}

func (v *Vault) isDir(dir *fuse.DirEntry) bool {
	name := dir.Name
	if name[len(name)-1:] == "/" {
		Log.Debug.Printf("isDir=true\n")
		return true
	}
	// if err is nil, then lookup in vault worked regularly
	// that means, it is a true directory in vault
	if _,err := v.listDir(name); err == nil {
		Log.Debug.Printf("isDir=true\n")
		return true
	}
	Log.Debug.Printf("isDir=false\n")
	return false
}


func (v *Vault) getType(name string) (Filetype, error){
	Log.Debug.Printf("op=getType name=\"%v\"\n",name)
	s,err := v.client.Logical().List(MTDATA + name)
	Log.Debug.Printf("op=getType s=\"%v\" err=\"%v\"\n",s,err)
	if err == nil && s != nil {
		return CTrueDir, nil
	}

	s,err = v.client.Logical().Read(DTDATA + name)
	if err == nil && s!=nil {
		return CFile, nil
	}

	name = path.Dir(name) // clip last element
	s,err = v.client.Logical().Read(DTDATA + name)
	if err == nil && s!=nil {
		return CValue, nil
	}

	return CNull, nil
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

