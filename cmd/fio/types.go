package fio

// FIOProvider interface provides all necessary calls used by FUSE
type FIOProvider interface {
       Open(string) error
       Read(string) error
}

// FIOMap maps the FIOProvider to a MountPath
type FIOMap struct {
       MountPath string
       Provider *FIOProvider
}
