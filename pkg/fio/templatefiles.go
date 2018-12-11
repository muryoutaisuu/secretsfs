package fio

import (
	"fmt"
	"os"
	"io/ioutil"
	"text/template"
	"bytes"
	"errors"
	"path"
	"strings"

	"github.com/muryoutaisuu/secretsfs/pkg/store"
	"github.com/hanwen/go-fuse/fuse"
	"github.com/hanwen/go-fuse/fuse/nodefs"
	"github.com/spf13/viper"
)

type FIOTemplatefiles struct {
	templpath string
}

type secret struct {
	flags uint32
	context *fuse.Context
	t *FIOTemplatefiles
}

func (t *FIOTemplatefiles) GetAttr(name string, context *fuse.Context) (*fuse.Attr, fuse.Status) {
	Log.Debug.Printf("ops=GetAttr name=\"%v\"\n",name)
	
	// opening directory (aka templatefiles/)
	if name == "" {
		return &fuse.Attr{
			Mode: fuse.S_IFDIR | 0550,
		}, fuse.OK
	}

	// get path to templates
	filepath := getCorrectPath(name)

	// check whether filepath exists
	file, err := os.Stat(filepath)
	if err != nil {
		Log.Error.Println(err)
		return nil, fuse.ENOENT
	}

	// get fileMode
	// https://stackoverflow.com/questions/8824571/golang-determining-whether-file-points-to-file-or-directory
	switch mode := file.Mode(); {
	case mode.IsDir():
		return &fuse.Attr{
			Mode: fuse.S_IFDIR | 0550,
		}, fuse.OK
	case mode.IsRegular():
		return &fuse.Attr{
			Mode: fuse.S_IFREG | 0550,
			Size: uint64(len(name)),
		}, fuse.OK
	}

	return nil, fuse.EINVAL
}

func (t *FIOTemplatefiles) OpenDir(name string, context *fuse.Context) ([]fuse.DirEntry, fuse.Status) {
	Log.Debug.Printf("ops=OpenDir name=\"%v\"\n",name)

	// get filepath to templates
	filepath := getCorrectPath(name)

	// check whether filepath exists
	file, err := os.Stat(filepath)
	if err != nil {
		Log.Error.Println(err)
		return nil, fuse.ENOENT
	}
	// check whether filepath is a directory
	// https://stackoverflow.com/questions/8824571/golang-determining-whether-file-points-to-file-or-directory
	if !file.Mode().IsDir() {
		Log.Error.Printf("op=OpenDir msg=\"not a directory\" filepath=\"%s\"\n",filepath)
		return nil, fuse.ENOTDIR
	}

	entries,err := ioutil.ReadDir(filepath)
	if err != nil {
		Log.Error.Print(err)
		return nil, fuse.EBUSY
	}
	dirs := []fuse.DirEntry{}
	for _,e := range entries {
		d := fuse.DirEntry{
			Name: e.Name(),
			Mode: uint32(e.Mode()),
		}
		dirs = append(dirs, d)
	}
	return dirs, fuse.OK
}

func (t *FIOTemplatefiles) Open(name string, flags uint32, context *fuse.Context) (nodefs.File, fuse.Status) {
	Log.Debug.Printf("ops=Open name=\"%v\"\n",name)

	// get filepath to templates
	filepath := getCorrectPath(name)

	// check whether filepath exists
	file, err := os.Stat(filepath)
	if err != nil {
		Log.Error.Println(err)
		return nil, fuse.ENOENT
	}

	// check whether filepath is a file
	// https://stackoverflow.com/questions/8824571/golang-determining-whether-file-points-to-file-or-directory
	if !file.Mode().IsRegular() {
		Log.Error.Printf("op=Open msg=\"not a directory\" filepath=\"%s\"\n",filepath)
		return nil, fuse.EISDIR
	}

	filename := path.Base(filepath)
	parser, err := template.New(filename).ParseFiles(filepath)
	// error handling
	if err != nil {
		errs := err.Error()
		Log.Error.Println(errs)
		return nil, fuse.EREMOTEIO
	}

	// https://gowalker.org/text/template#Template_Execute
	// https://yourbasic.org/golang/io-writer-interface-explained/
	// https://gowalker.org/bytes#Buffer_Bytes
	// https://stackoverflow.com/questions/23454940/getting-bytes-buffer-does-not-implement-io-writer-error-message
	var buf bytes.Buffer
	secret := secret{
		flags: flags,
		context: context,
		t: t,
	}

	err = parser.Execute(&buf, secret)
	if err != nil {
		Log.Error.Println(err)
		switch {
		case strings.Contains(err.Error(), fmt.Sprint(fuse.EACCES)):
			return nil, fuse.EACCES
		default:
			return nil, fuse.EREMOTEIO
		}
	}

	return nodefs.NewDataFile(buf.Bytes()), fuse.OK
}

func (t *FIOTemplatefiles) FIOPath() string {
	return "templatefiles"
}

func getCorrectPath(name string) string {
	filepath := viper.GetString("PATH_TO_TEMPLATES")+name
	Log.Debug.Printf("op=getCorrectPath variable=filepath value=\"%s\"\n",filepath)
	return filepath
}

func (s secret) Get(filepath string) (string, error) {
	sto := store.GetStore()
  content, status := sto.Open(filepath, s.flags, s.context)
	if status != fuse.OK {
		Log.Error.Printf("op=Get msg=\"There was an error while loading secret from store\" fuse.Status=\"%s\"\n",status)
		//return "", errors.New("There was an error while loading Secret from store, fuse.Status="+fmt.Sprint(status))
		return "", errors.New(fmt.Sprint(status))
	}
	return content, nil
}



func init() {
  secretsfiles := FIOSecretsfiles{}
  fm := FIOMap{
    Provider: &secretsfiles,
  }
  RegisterProvider(&fm)
}



func init() {
	fioprov := FIOTemplatefiles{
		templpath: viper.GetString("PATH_TO_TEMPLATES"),
	}
	fm := FIOMap{
		Provider: &fioprov,
	}
	RegisterProvider(&fm)
}
