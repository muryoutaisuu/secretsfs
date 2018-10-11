package store

import (
	"io/ioutil"
	"path/filepath"
	"os/user"
	"strings"

	"github.com/hashicorp/vault/api"
)


type Vault struct {
	client *api.Client
}

func (v *Vault) List(u *user.User, path string) error {
	return nil

}

func (v *Vault) Read(u *user.User, path string) error {
	return nil
}

func (v *Vault) Write(u *user.User, path,content string) error {
	return nil
}

func (v *Vault) Delete(u *user.User, path string) error {
	return nil
}

func (v *Vault) String() (string, error) {
	return "Vault",nil
}

func (v *Vault) Client() (*api.Client, error) {
	return v.client, nil
}




func (v *Vault) secret(u *user.User) (*api.Secret, error) {
	authToken,_ := readAuthToken(u)
	c,_ := v.Client()
	auth := c.Auth()
	tokenauth := auth.Token()
	secret,err := tokenauth.Lookup(authToken)
	return secret,err
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

