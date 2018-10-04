package fio

type FIOTemplatefiles struct {}

func (t *FIOTemplatefiles) Open(file string) error {
       return nil
}

func (t *FIOTemplatefiles) Read(file string) error {
       return nil
}

func init() {
       provider := FIOMap {
             MountPath: "templatefiles",
             Provider: &FIOTemplatefiles{},
       }
       
       fio.RegisterProvider(provider)
}
