package store

type Vault struct {}

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

func init() {
	v := Vault{}
	RegisterStore(&v)
}
