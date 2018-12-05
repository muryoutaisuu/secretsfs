package fio

import (
	"os"
	"io/ioutil"
	"text/template"
	"bytes"

	"github.com/Muryoutaisuu/secretsfs/pkg/store"
	"github.com/hanwen/go-fuse/fuse"
	"github.com/hanwen/go-fuse/fuse/nodefs"
	"github.com/spf13/viper"
)

type FIOTemplatefiles struct {
	path  string
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
	path := getCorrectPath(name)

	// check whether path exists
	file, err := os.Stat(path)
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

	// get path to templates
	path := getCorrectPath(name)

	// check whether path exists
	file, err := os.Stat(path)
	if err != nil {
		Log.Error.Println(err)
		return nil, fuse.ENOENT
	}
	// check whether path is a directory
	// https://stackoverflow.com/questions/8824571/golang-determining-whether-file-points-to-file-or-directory
	if !file.Mode().IsDir() {
		Log.Error.Printf("op=OpenDir msg=\"not a directory\" path=\"%s\"\n",path)
		return nil, fuse.ENOTDIR
	}

	entries,err := ioutil.ReadDir(path)
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

	// get path to templates
	path := getCorrectPath(name)

	// check whether path exists
	file, err := os.Stat(path)
	if err != nil {
		Log.Error.Println(err)
		return nil, fuse.ENOENT
	}

	// check whether path is a file
	// https://stackoverflow.com/questions/8824571/golang-determining-whether-file-points-to-file-or-directory
	if !file.Mode().IsRegular() {
		Log.Error.Printf("op=Open msg=\"not a directory\" path=\"%s\"\n",path)
		return nil, fuse.ENOTDIR
	}
	
	// read template
	templ, err := ioutil.ReadFile(path)
	if err != nil {
		Log.Error.Print(err)
		return nil, fuse.EIO
	}

	templs := string(templ)
	parser, err := template.New("Open").Parse(templs)
	if err != nil {
		Log.Error.Println(err)
		return nil, fuse.EIO
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
		return nil, fuse.EIO
	}

	return nodefs.NewDataFile(buf.Bytes()), fuse.OK
}

func getCorrectPath(name string) string {
	path := viper.GetString("PATH_TO_TEMPLATES")+name
	Log.Debug.Printf("op=getCorrectPath variable=path value=\"%s\"\n",path)
	return path
}

func (s secret) Get(path string) string {
	sto := store.GetStore()
  content, _ := sto.Open(path, s.flags, s.context)
	return content
}




func init() {
	name := "templatefiles"
	fios := viper.GetStringSlice("ENABLED_FIOS")
	for _,f := range fios {
		if f == name {
			templatefiles := FIOTemplatefiles{
				path: viper.GetString("PATH_TO_TEMPLATES"),
			}
			fm := FIOMap {
				MountPath: name,
				Provider: &templatefiles,
			}

			RegisterProvider(&fm)
		}
	}
}
