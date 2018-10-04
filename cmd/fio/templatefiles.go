package fio

type FIOTemplatefiles struct {}

func (t *FIOTemplatefiles) Open(file string) error {
       return nil
}

func (t *FIOTemplatefiles) Read(file string) error {
       return nil
}

func init() {
       fm := FIOMap {
             MountPath: "templatefiles",
             Provider: &FIOTemplatefiles{},
       }
       
       RegisterProvider(&fm)
}
