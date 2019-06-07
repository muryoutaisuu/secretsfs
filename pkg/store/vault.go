package store

import (
	"errors"
	"strconv"
	"io/ioutil"
	"os/user"
	"strings"
	"path"
	"path/filepath"

	"github.com/hashicorp/vault/api"
	"github.com/hanwen/go-fuse/fuse"
	"github.com/spf13/viper"
	log "github.com/sirupsen/logrus"

	sfsh "github.com/muryoutaisuu/secretsfs/pkg/sfshelpers"
)

// Path internals of vault made configurable with viper
// taken from https://www.vaultproject.io/api/secret/kv/kv-v2.html
//var (
//	MTDATA string
//	DTDATA string
//)
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
// It's a store and may be coupled with multiple fio structs
type Vault struct {
	client *api.Client
}

func (v *Vault) GetAttr(name string, context *fuse.Context) (*fuse.Attr, fuse.Status) {
	u,err := sfsh.GetUser(context)
	if err != nil {
		return nil, fuse.EPERM
	}
	logger = defaultEntry(name, u)
	logger.Info("calling operation")

	// opening directory (aka secretsfiles/)
	if name == "" {
		return &fuse.Attr{
			Mode: fuse.S_IFDIR | 0550,
		}, fuse.OK
	}

	if err := v.setToken(context); err != nil {
		logger.Error(err)
		return nil, fuse.EACCES
	}
	defer logger.Debug("successfully cleared token")
	defer v.client.ClearToken()

	// get type
	_, t := v.getTypes(name)
	logger.WithFields(log.Fields{"types":t}).Debug("got types")

	// act according to type
	switch {
	case t[CTrueDir], t[CFile]:
		return &fuse.Attr{
			Mode: fuse.S_IFDIR | 0550,
		}, fuse.OK
	case t[CValue]:
		return &fuse.Attr{
			Mode: fuse.S_IFREG | 0550,
			Size: uint64(len(name)),
		}, fuse.OK
	default: // probably not enough permissions to determine type -> would probably be a directory
		return &fuse.Attr{
			Mode: fuse.S_IFDIR | 0000,
		}, fuse.OK
	}
}

func (v *Vault) OpenDir(name string, context *fuse.Context) ([]fuse.DirEntry, fuse.Status) {
	u,err := sfsh.GetUser(context)
	if err != nil {
		return nil, fuse.EPERM
	}
	logger = defaultEntry(name, u)
	logger.Info("calling operation")

	if err := v.setToken(context); err != nil {
		logger.Error(err)
		return nil, fuse.EACCES
	}
	defer logger.Debug("successfully cleared token")
	defer v.client.ClearToken()

	_, t := v.getTypes(name)
	logger.WithFields(log.Fields{"types":t}).Debug("got types")

	finuniqnames := make(map[string]struct{})
	p := &finuniqnames
	logger.WithFields(log.Fields{"finuniqnames":finuniqnames}).Debug("got finuniqnames")
	if t[CTrueDir] {
		err := v.listDirUniqueNames(name, p)
		if err != nil && !t[CFile] {
			logger.Error(err)
			return nil, fuse.EIO
		}
	}
	if t[CFile] {
		err := v.listFileUniqueNames(name, p)
		if err != nil && !t[CTrueDir] {
			logger.Error(err)
			return nil, fuse.EIO
		}
	}
	if t[CValue] {
		return nil, fuse.ENOTDIR
	}
	if len(*p) > 0 {
		findirs := []fuse.DirEntry{}
		for k := range *p {
			d := fuse.DirEntry{
				Name: k,
				Mode: fuse.S_IFREG,
			}
			findirs = append(findirs,d)
		}
		return findirs, fuse.OK
	}
	return nil, fuse.ENOENT
}

func (v *Vault) Open(name string, flags uint32, context *fuse.Context) (string, fuse.Status) {
	u,err := sfsh.GetUser(context)
	if err != nil {
		return "", fuse.EPERM
	}
	logger = defaultEntry(name, u)
	logger.Info("calling operation")

	if err := v.setToken(context); err != nil {
		logger.Error(err)
		return "", fuse.EACCES
	}
	defer logger.Debug("successfully cleared token")
	defer v.client.ClearToken()

	s,t := v.getTypes(name)
	logger.WithFields(log.Fields{"types":t}).Debug("got types")

	switch {
	case t[CTrueDir]:
		return "", fuse.EISDIR
	case t[CFile]:
		return "", fuse.EISDIR
	case t[CValue]:
		// get substituted value (if substitution must be done, else keep original)
		logger.WithFields(log.Fields{"variable":"name","value":name}).Debug("before substituting")
		name, _, err := v.getCorrectName(name, true)
		if err != nil {
			logger.Error(err)
			return "", fuse.EIO
		}
		logger.WithFields(log.Fields{"variable":"name","value":name}).Debug("after substituting")

		logger.WithFields(log.Fields{"s[CValue]":s[CValue]}).Debug("log values")
		data,ok := s[CValue].Data[name].(string)
		if ok != true {
			return "", fuse.EIO
		}
		return data, fuse.OK
	}
	return "", fuse.ENOENT
}

func (v *Vault) String() (string) {
	return "vault"
}




// setToken is called within the fuse interaction calls and sets a working
// accesstoken depending on the calling user
// usually should be used in conjunction to a deferred clear call:
// if err := v.setToken(context); err != nil {
// 	logger.Error(err)
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
	logger.WithFields(log.Fields{"token":v.client.Token()}).Debug("log values")
	return nil
}

// getAccessToken reads the currently set authentication token inside of the
// users home and authenticates with it and returns afterwards the secret
// containing the accesstoken
func (v *Vault) getAccessToken(u *user.User) (*api.Secret, error) {
	auth,err := v.readAuthToken(u)
	if err != nil {
		logger.Error(err)
		return &api.Secret{}, err
	}
	// https://groups.google.com/forum/#!topic/vault-tool/-4F2RLnGrSE
	postdata := map[string]interface{}{
		"role_id": auth,
	}
	logger.WithFields(log.Fields{"login_payload":postdata}).Debug("authenticating")
	resp,err := v.client.Logical().Write("auth/approle/login", postdata)
	if err != nil {
		logger.Error("got an error while authenticating")
		return nil,err
	}
	logger.WithFields(log.Fields{"resp":resp, "resp.Data":resp.Data, "ClientToken":resp.Auth.ClientToken}).Debug("log values")
	if resp.Auth == nil {
		return resp, errors.New("no auth info returned")
	}
	logger.Debug("successfully got accesstoken")
	return resp,err
}

// readAuthToken opens the file containing the authenticationtoken and trims it
func (v *Vault) readAuthToken(u *user.User) (string, error) {
	path := finIdPath(u)
	logger.WithFields(log.Fields{"finIdPath":path}).Debug("log values")
	o,err := ioutil.ReadFile(path)
	if err != nil {
		logger.Error(err)
		return "",err
	}
	authToken := strings.TrimSuffix(string(o), "\n")
	logger.WithFields(log.Fields{"finIdPath":path}).Debug("authtoken successfully read")
	return authToken,nil
}

// listDir lists all entries inside a vault directory type=CTrueDir
func (v *Vault) listDir(name string) (*[]fuse.DirEntry, error) {
	s,err := v.listDirNames(name)
	if err != nil {
		logger.WithFields(log.Fields{"secret":s, "error":err}).Error("got an error in listDirNames")
		return nil, errors.New("Got an error in listDirNames")
	}
	logger.WithFields(log.Fields{"secret":s}).Debug("log values")
	dirs := []fuse.DirEntry{}
	logger.WithFields(log.Fields{"dirs":dirs}).Debug("log values")
	for i := 0; i < len(s); i++ {
		d := fuse.DirEntry{
			Name:  path.Base(s[i]),
			Mode: fuse.S_IFREG,
		}
		dirs = append(dirs, d)
		logger.WithFields(log.Fields{"dirs":dirs}).Debug("log values")
	}
	return &dirs,nil
}

// listDirNames lists all entries inside a vault directory type=CTrueDir and
// returns a []string with the names of those directories
func (v *Vault) listDirNames(name string) ([]string, error) {
	logger.WithFields(log.Fields{"url":v.client.Address()+MTDATA+name}).Debug("listing directory in vault")
	s,err := v.client.Logical().List(MTDATA + name+"/")
	logger.WithFields(log.Fields{"url":v.client.Address()+MTDATA+name, "secret":s}).Debug("log values")

	// can't list in vault
	if err != nil || s == nil {
		if err == nil {
			err = errors.New("cant list path "+MTDATA+name+" in vault")
		}
		logger.Error(err)
		return nil, err
	}

	names := []string{}
	for _,v := range s.Data["keys"].([]interface{}) {
		names = append(names, v.(string))
	}
	return names, nil
}

func (v *Vault) listDirUniqueNames(name string, un *map[string]struct{}) (error) {
	if un == nil {
		return errors.New("nil is not a supported value for un")
	}
	names, err := v.listDirNames(name)
	if err != nil {
		return errors.New("Got an error in listDirNames")
	}

	for _,v := range names {
		(*un)[filepath.Base(v)] = struct{}{}
	}
	return nil
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

	dirs := v.createFileEntries(data)
	logger.WithFields(log.Fields{"dirs":dirs}).Debug("log values")
	return dirs,nil
}

func (v *Vault) createFileEntries(names []string) (dirs *[]fuse.DirEntry) {
	for _,v := range names {
		// special treatment for entries containing the substitution character
		if strings.Contains(v, "/") { // viper.GetString("general.substchar")) { // strings.Contains(k,"/") {
			v = strings.Replace(v, "/", string(viper.GetString("general.substchar")[0]), -1)
		}

		d := fuse.DirEntry{
			Name: v,
			Mode: fuse.S_IFREG,
		}
		*dirs = append(*dirs, d)
	}
	return dirs
}

// listFileNames is very similar to listFile, but instead of returning fully
// finished fuse.DirEntry types, it only returns []string containing the keys
func (v *Vault) listFileNames(name string) ([]string, error) {
	logger.WithFields(log.Fields{"url":v.client.Address()+DTDATA+name}).Debug("reading secret in vault")
	s,err := v.client.Logical().Read(DTDATA + name)
	if err != nil || s == nil {
		if err == nil {
			logger.WithFields(log.Fields{"url":v.client.Address()+DTDATA+name}).Error("can't read secret")
			errors.New("can't read")
		}
		return nil,err
	}
	logger.WithFields(log.Fields{"secret":s, "secret.Data":s.Data}).Debug("log values")

  filenames := []string{}
	for k := range s.Data {
		filenames = append(filenames, k)
	}
	return filenames, nil
}

func (v *Vault) listFileUniqueNames(name string, un *map[string]struct{}) (error) {
	if un == nil {
		return errors.New("nil is not a supported value for un")
	}
	names, err := v.listFileNames(name)
	if err != nil {
		return errors.New("Got an error in listFileNames")
	}

	for _,v := range names {
		(*un)[v] = struct{}{}
	}
	return nil
}

// getType returns type of the requested resource
// used by most fuse actions for simplifying reasons
// types may be the defined FileType byte constants on top of this file
func (v *Vault) getType(name string) (*api.Secret, Filetype){
	logger.WithFields(log.Fields{"url":v.client.Address()+MTDATA+name}).Debug("listing directory in vault")
	s,err := v.client.Logical().List(MTDATA + name + "/")
	logger.WithFields(log.Fields{"url":v.client.Address()+MTDATA+name, "secret":s, "error":err}).Error("after listing directory in vault")
	if err == nil && s != nil {
		return s, CTrueDir
	}

	logger.WithFields(log.Fields{"url":v.client.Address()+DTDATA+name}).Debug("reading secret in vault")
	s,err = v.client.Logical().Read(DTDATA + name)
	if err == nil && s!=nil {
		return s, CFile
	}

	name = path.Dir(name) // clip last element
	logger.WithFields(log.Fields{"url":v.client.Address()+DTDATA+name}).Debug("reading secret in vault")
	s,err = v.client.Logical().Read(DTDATA + name)
	if err == nil && s!=nil {
		return s, CValue
	}

	return nil, CNull
}

// getTypes returns similar to getType the types of the requested resources
// imagine following situation:
//   secret/foo
//   secret/foo/
//   secret/foo/bar
// here foo is a secret as well as a subdirectory. It should be possible, to
// get both those types
func (v *Vault) getTypes(name string) (map[Filetype]*api.Secret, map[Filetype]bool) {
	r := make(map[Filetype]bool)
	rs := make(map[Filetype]*api.Secret)

	logger.WithFields(log.Fields{"url":v.client.Address()+MTDATA+name}).Debug("listing directory in vault")
	s,err := v.client.Logical().List(MTDATA + name + "/")
	logger.WithFields(log.Fields{"url":v.client.Address()+MTDATA+name, "secret":s, "error":err}).Debug("after listing directory in vault")
	r[CTrueDir] = err == nil && s != nil
	rs[CTrueDir] = s

	logger.WithFields(log.Fields{"url":v.client.Address()+DTDATA+name}).Debug("reading secret in vault")
	s,err = v.client.Logical().Read(DTDATA + name)
	logger.WithFields(log.Fields{"url":v.client.Address()+DTDATA+name, "secret":s, "error":err}).Debug("after reading secret in vault")
	r[CFile] = err == nil && s != nil
	rs[CFile] = s

	// if else statement here is needed, case of:
	//   E1							->  secret
	//   E1/mysecret = 42
	//   E1/						->  subdir in Vault
	//   E1/subsecret		-> secret
	//   E1/subsecret/mysecret = 43
	// this would have thrown an error, because for E1/subsecret/mysecret it would
	// have r[CFile] == true AND r[CValue] == true
	// this would cause errors in any further calculations
	if !r[CFile] {
		name = path.Dir(name) // clip last element
		logger.WithFields(log.Fields{"url":v.client.Address()+DTDATA+name}).Debug("reading secret in vault")
		s,err = v.client.Logical().Read(DTDATA + name)
		logger.WithFields(log.Fields{"url":v.client.Address()+DTDATA+name, "secret":s, "error":err}).Debug("after reading secret in vault")
		r[CValue] = err == nil && s != nil
		rs[CValue] = s
	} else {
		r[CValue] = false
		rs[CValue] = nil
	}

	return rs,r
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
	if !strings.Contains(value, viper.GetString("general.substchar")) {
		logger.WithFields(log.Fields{"variable":"value","value":value}).Debug("contains no characters that may be substituted")
		return value, false, nil
	}

	dir := path.Dir(pathname)
	logger.WithFields(log.Fields{"variable":"dir","value":dir}).Debug("doing a listFileNames with specific dir")
	filenames,err := v.listFileNames(dir)
	if err != nil {
		logger.WithFields(log.Fields{"variable":"dir","value":dir, "error":err}).Debug("got an error doing a listFileNames with specific dir")
		return "", false, err
	}
	logger.WithFields(log.Fields{"variable":"filenames","value":filenames}).Debug("got actual contents of vault secret")

	possibilities := sfsh.SubstitutionPossibilities(value, viper.GetString("general.substchar"), "/")
	logger.WithFields(log.Fields{"variable":"possibilities", "value":possibilities}).Debug("got all possible key names")
	for _,f := range filenames {
		for _,p := range possibilities {
			if f == p {
				logger.WithFields(log.Fields{"variable":"filename", "value":f}).Debug("log values")
				if nameonly {
					return f, true, nil
				}
				return dir+"/"+f, true, nil
			}
		}
	}
	logger.WithFields(log.Fields{"variable":"possibilities", "value":possibilities}).Error("can't find any substituted possibilities")
	return "", false, errors.New("can't find any substituted possibilties for value "+value)
}

// finIdPath returns the resolved path of the users vault roleid file path.
// This means that $HOME will be resolved to the users home directory, and that
// the users alias is applied
func finIdPath(u *user.User) (string) {
	path := strings.Replace(viper.GetString("store.vault.roleid.file"), "$HOME", u.HomeDir, 1)

	specialusers := viper.GetStringMapString("store.vault.roleid.useroverride")
	if val, ok := specialusers[u.Name]; ok {
		// replace $HOME also, if path was set user specific
		path = strings.Replace(val, "$HOME", u.HomeDir, 1)
	}
	return path
}

func configureTLS(c *api.Config) error {
	tls := api.TLSConfig{}
	if viper.IsSet("tls.cacert") { tls.CACert = viper.GetString("tls.cacert") }
	if viper.IsSet("tls.capath") { tls.CAPath = viper.GetString("tls.capath") }
	if viper.IsSet("tls.clientcert") { tls.ClientCert = viper.GetString("tls.clientcert") }
	if viper.IsSet("tls.clientkey") { tls.ClientKey = viper.GetString("tls.clientkey") }
	if viper.IsSet("tls.tlsservername") { tls.TLSServerName = viper.GetString("tls.tlsservername") }
	if viper.IsSet("tls.insecure") { tls.Insecure = viper.GetBool("tls.insecure") }
	err :=  c.ConfigureTLS(&tls)
	if c.Error != nil { return c.Error }
	return err
}

func init() {
	a := viper.GetString("store.vault.addr")
	// create first config type
	conf := api.DefaultConfig()
	conf.Address = a

	// check whether TLS is needed
	if len(a) >= 5 && a[:5] == "https" {
		if err := configureTLS(conf); err != nil {
			logger.Fatal(err)
		}
	}

	// create client
	c,err := api.NewClient(conf)
	if err != nil {
		logger.Fatal(err)
	}

	// create vault object & register it
	v := Vault{
		client: c,
	}
	v.client.ClearToken()
	RegisterStore(&v) //https://stackoverflow.com/questions/40823315/x-does-not-implement-y-method-has-a-pointer-receiver
	if viper.GetString("store.enabled") == v.String() {
		MTDATA = viper.GetString("store.vault.mtdata")
		DTDATA = viper.GetString("store.vault.dtdata")
	}
}

