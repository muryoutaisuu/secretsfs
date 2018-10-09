package fio

type FIOSecretsfiles struct {}

func (t *FIOSecretsfiles) Open(file string) error {
       return nil
}

func (t *FIOSecretsfiles) Read(file string) error {
       return nil
}

func init() {
       fm := FIOMap {
             MountPath: "secretsfiles",
             Provider: &FIOSecretsfiles{},
       }
       
       RegisterProvider(&fm)
}
