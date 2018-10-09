package store

import (
	"io/ioutil"
	"path/filepath"
	"os/user"
	"strings"

	api "github.com/hashicorp/vault/api"
)


// Vault FS Interaction things

type Vault struct {
	client *api.Client
}

func (v *Vault) List(path string) error {
	return nil
}

func (v *Vault) Read(path string) error {
	return nil
}

func (v *Vault) Write(path,content string) error {
	return nil
}

func (v *Vault) Delete(path string) error {
	return nil
}

func (v *Vault) String() (string, error) {
	return "Vault",nil
}


// authBundle things

// authBundle contains all Tokens of Vault and the user struct
type authBundle struct{
	authtoken string
	acctoken string
	user *user.User
}

func (a *authBundle) RenewAccToken() error {
	return nil
}

func (a *authBundle) AccToken() (string, error) {
	return a.acctoken, nil
}

func AuthBundle(u *user.User) (*authBundle, error) {
	aut,_ := readAuthToken(u)
	return &authBundle{
		authtoken: aut,
		acctoken: "",
		user: u,},
	nil
}

func readAuthToken(u *user.User) (string, error) {
	path := filepath.Join(u.HomeDir, "/authTokenfile")
	Linf.Println(path)
	o,err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}
	authToken := strings.TrimSuffix(string(o), "\n")
	return authToken,nil
}





func init() {
	c,_ := api.NewClient(&api.Config{
		Address: "",
	})
	v := Vault{
		client: c,
	}
	RegisterStore(&v) //https://stackoverflow.com/questions/40823315/x-does-not-implement-y-method-has-a-pointer-receiver
}

