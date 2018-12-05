package store

import (
	"errors"
	"fmt"
	"strconv"
	"io/ioutil"
	"path/filepath"
	//"os"
	"os/user"
	"strings"
	"path"
	//"encoding/json"

	//"gopkg.in/yaml.v2"

	"github.com/hashicorp/vault/api"
	"github.com/hanwen/go-fuse/fuse"
	"github.com/spf13/viper"

	"github.com/Muryoutaisuu/secretsfs/pkg/sfshelpers"
)

// Path internals of vault made configurable with viper
// taken from https://www.vaultproject.io/api/secret/kv/kv-v2.html
var MTDATA string
var DTDATA string

// Filetype define the type of the returned value element of vault
type Filetype byte
const (
	CTrueDir   Filetype = 0 // exists in Vault as a directory
	CFile      Filetype = 1 // Key of a key=value pair, emulated as a directory
	CValue     Filetype = 2 // Value of a key=value pair
	CNull      Filetype = 3 // not a valid vault element
)


// Vault struct implements the calls called by fuse and returns accordingly
// requested resources.
// it's a store and may be coupled with multiple fio structs
type Vault struct {
	client *api.Client
	//TokenAuth *api.Client.Auth().Token()
}

func (v *Vault) GetAttr(name string, context *fuse.Context) (*fuse.Attr, fuse.Status) {
	Log.Debug.Printf("ops=GetAttr name=\"%v\"\n",name)
	Log.Debug.Printf("ops=GetAttr MTDATA=%s",viper.GetString("MTDATA"))
	Log.Debug.Printf("ops=GetAttr Token=%s",v.client.Token())
	//name = MTDATA + name

	// opening directory (aka secretsfiles/)
	if name == "" {
		return &fuse.Attr{
			Mode: fuse.S_IFDIR | 0550,
		}, fuse.OK
	}

	if err := v.setToken(context); err != nil {
		Log.Error.Print(err)
		return nil, fuse.EACCES
	}
	defer Log.Debug.Printf("op=GetAttr msg=\"successfully cleared token\" token=%s\"\n",v.client.Token())
	defer v.client.ClearToken()
	defer Log.Debug.Printf("op=GetAttr msg=\"successfully cleared token\" token=%s\"\n",v.client.Token())

	// get type
	Log.Debug.Printf("name=\"%v\"\n",name)
	_,t := v.getType(name)
	Log.Debug.Printf("op=GetAttr t=\"%v\"\n",t)

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

	if err := v.setToken(context); err != nil {
		Log.Error.Print(err)
		return nil, fuse.EACCES
	}
	defer Log.Debug.Printf("op=OpenDir msg=\"successfully cleared token\" token=%s\"\n",v.client.Token())
	defer v.client.ClearToken()

	_,t := v.getType(name)
	Log.Debug.Printf("ops=OpenDir t=\"%v\"\n",t)

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

func (v *Vault) Open(name string, flags uint32, context *fuse.Context) (string, fuse.Status) {
	Log.Debug.Printf("op=Open name=\"%v\"\n",name)

	if err := v.setToken(context); err != nil {
		Log.Error.Print(err)
		return "", fuse.EACCES
	}
	defer Log.Debug.Printf("op=Open msg=\"successfully cleared token\" token=%s\"\n",v.client.Token())
	defer v.client.ClearToken()

	s,t := v.getType(name)
	Log.Debug.Printf("op=Open t=\"%v\"\n",t)

	switch t {
	case CTrueDir:
		return "", fuse.EISDIR
	case CFile:
		return "", fuse.EISDIR
	case CValue:
		// get substituted value (if substitution must be done, else keep original)
		Log.Debug.Printf("op=Open msg=\"before substituting name\" variable=name value=%v\n",name)
		name, _, err := v.getCorrectName(name, true)
		if err != nil {
			Log.Error.Print(err)
			return "", fuse.EIO
		}
		Log.Debug.Printf("op=Open msg=\"after substituting name\" variable=name value=%v\n",name)

		Log.Debug.Printf("op=Open s=\"%v\" name=\"%v\"\n",s,name)
		data,ok := s.Data["data"].(map[string]interface{})
		if ok != true {
			return "", fuse.EIO
		}
		entry,ok := data[name].(string)
		if ok != true {
			return "", fuse.EIO
		}
		return entry, fuse.OK
	}
	return "", fuse.ENOENT
}

func (v *Vault) String() (string) {
	return "Vault"
}




// setToken is called within the fuse interaction calls and sets a working
// accesstoken depending on the calling user
// usually should be used in conjunction to a deferred clear call:
// if err := v.setToken(context); err != nil {
// 	Log.Error.Print(err)
// 	return nil, fuse.EACCES
// }
// defer v.client.ClearToken()
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
	// TODO: Remove this debug line, not secure!!
	Log.Debug.Printf("op=setToken msg=\"successfully set token\" token=%s\"\n",v.client.Token())
	return nil
}

// getAccessToken reads the currently set authentication token inside of the
// users home and authenticates with it and returns afterwards the secret
// containing the accesstoken
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
	if err != nil {
		Log.Error.Printf("op=getAccessToken msg=\"Got an error while authenticating\"\n")
		return nil,err
	}
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

// readAuthToken opens the file containing the authenticationtoken and trimps it
func (v *Vault) readAuthToken(u *user.User) (string, error) {
	// path := filepath.Join(u.HomeDir, os.Getenv("SECRETSFS_FILE_ROLEID"))
	path := filepath.Join(u.HomeDir, viper.GetString("FILE_ROLEID"))
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

// listDir lists all entries inside a vault directory type=CTrueDir
func (v *Vault) listDir(name string) (*[]fuse.DirEntry, error) {
	Log.Debug.Printf("op=listDir MTDATA=\"%v\" name=\"%v\"",MTDATA,name)
	s,err := v.client.Logical().List(MTDATA + name)
	Log.Debug.Printf("secret=\"%v\"\n",s)

	// can't list in vault
	if err != nil || s == nil {
		if err == nil {
			err = errors.New("cant list path "+MTDATA+name+" in vault")
		}
		Log.Error.Print(err)
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

// listFile lists the contents of a virtual directory in secretsfs
// (aka a file in vault) type=CFile
// returns a Slice containing all valid entries
// valid means no entries containing a / in their names
func (v *Vault) listFile(name string) (*[]fuse.DirEntry, error) {
	data,err := v.listFileNames(name)
	if err != nil {
		return nil, err
	}

	dirs := []fuse.DirEntry{}
	for k := range data {
		key := data[k]
		// special treatment for entries containing the substitution character
		if strings.Contains(key, "/") { // viper.GetString("subst_char")) { // strings.Contains(k,"/") {
			key = strings.Replace(key, "/", string(viper.GetString("subst_char")[0]), -1)
		}

		d := fuse.DirEntry{
			Name: key,
			Mode: fuse.S_IFREG,
		}
		dirs = append(dirs, d)
	}
	Log.Debug.Printf("op=listFile dirs=\"%v\"\n",dirs)
	return &dirs,nil
}

// listFileNames is very similar to listFile, but instead of returning fully
// finished fuse.DirEntry types, it only returns []string containing the keys
func (v *Vault) listFileNames(name string) ([]string, error) {
	Log.Debug.Printf("op=listFileNames msg=\"going to read data\" path=\"%s\"\n",DTDATA + name)
	s,err := v.client.Logical().Read(DTDATA + name)
	if err != nil || s == nil {
		if err == nil {
			errors.New("cant read")
		}
		return nil,err
	}
	Log.Debug.Printf("op=listFile secret=\"%v\"\n",s)
	Log.Debug.Printf("op=listFile secret.Data=\"%v\" secret.DataType=\"%T\"\n",s.Data,s.Data)
	data,ok := s.Data["data"].(map[string]interface{})
	if !ok {
		return nil, errors.New("s.Data[\"data\"] resulted in a error")
	}
	Log.Debug.Printf("op=listFileNames data=\"%v\" dataType=\"%T\"\n",data,data)

  filenames := []string{}
	for k := range data {
		filenames = append(filenames, k)
	}
	return filenames, nil
}

// getType returns type of the requested resource
// used by most fuse actions for simplifying reasons
// types may be the defined FileType byte constants on top of this file
func (v *Vault) getType(name string) (*api.Secret, Filetype){
	Log.Debug.Printf("op=getType name=\"%v\"\n",name)
	s,err := v.client.Logical().List(MTDATA + name)
	Log.Debug.Printf("op=getType MTDATA=%s",MTDATA)
	Log.Debug.Printf("op=getType s=\"%v\" err=\"%v\"\n",s,err)
	if err == nil && s != nil {
		return s, CTrueDir
	}

	s,err = v.client.Logical().Read(DTDATA + name)
	if err == nil && s!=nil {
		return s, CFile
	}

	name = path.Dir(name) // clip last element
	s,err = v.client.Logical().Read(DTDATA + name)
	if err == nil && s!=nil {
		return s, CValue
	}

	return nil, CNull
}

// getCorrectName checks whether a path contains any maybe substituted characters.
// If yes, it checks in Vault whether there is a substituted key available and
// returns it incl. whole path.
// if only the Name itself is wished, set nameonly=true
// if no value found, then throws an error and returns ""
//
// Why is there a nameonly=true ?
// Problem being the fact, that with the original value in Vault the correct
// path for getting the Secret from Vault may be quite tricky
// e.g. the substituted value:  GET secret/my_bad_key
// would become:                GET secret/my/bad/key
// where a simple path.Base(path) won't return the secret's name anymore
func (v *Vault) getCorrectName(pathname string, nameonly bool) (string, bool, error) {
	value := pathname

	// split if nameonly is true
	if nameonly {
		value = path.Base(pathname)
	}

	// check whether name contains any characters, that may be substituted
	if !strings.Contains(value, viper.GetString("subst_char")) {
		Log.Debug.Printf("op=getCorrectName msg=\"contains no characters that may be substituted\" variable=value value=\"%v\"\n",value)
		return value, false, nil
	}

	dir := path.Dir(pathname)
	Log.Debug.Printf("op=getCorrectName msg=\"do a listFileNames with specific dir\" variable=dir value=\"%v\"\n",dir)
	filenames,err := v.listFileNames(dir)
	if err != nil {
		return "", false, err
	}
	Log.Debug.Printf("op=getCorrectName msg=\"got actual contents of vault secret\" variable=filenames value=\"%v\"\n",filenames)

	possibilities := sfshelpers.SubstitutionPossibilities(value, viper.GetString("subst_char"), "/")
	Log.Debug.Printf("op=getCorrectName msg=\"got all possible key names\" variable=possibilities value=\"%v\"\n",possibilities)
	for _,f := range filenames {
		for _,p := range possibilities {
			if f == p {
				Log.Debug.Printf("op=getCorrectName msg=\"found correct substituted name\" variable=filename value=\"%v\"\n",f)
				if nameonly {
					return f, true, nil
				}
				return dir+"/"+f, true, nil
			}
		}
	}
	return "", false, errors.New("can't find any substituted possibilties for value "+value)
}





func init() {
	c,err := api.NewClient(&api.Config{
		// Address: os.Getenv("VAULT_ADDR"),
		Address: viper.GetString("VAULT_ADDR"),
	})
	if err != nil {
		Log.Error.Fatal(err)
	}
	v := Vault{
		client: c,
	}
	v.client.ClearToken()
	RegisterStore(&v) //https://stackoverflow.com/questions/40823315/x-does-not-implement-y-method-has-a-pointer-receiver
	if viper.GetString("CURRENT_STORE") == v.String() {
		Log.Debug.Printf("op=init MTDATA=%s",viper.GetString("MTDATA"))
		MTDATA = viper.GetString("MTDATA")
		DTDATA = viper.GetString("DTDATA")
	}
}

