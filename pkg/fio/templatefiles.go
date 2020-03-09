package fio

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"text/template"

	sfsh "github.com/muryoutaisuu/secretsfs/pkg/sfshelpers"
	sfsl "github.com/muryoutaisuu/secretsfs/pkg/sfslog"
	"github.com/muryoutaisuu/secretsfs/pkg/store"

	"github.com/hanwen/go-fuse/fuse"
	"github.com/hanwen/go-fuse/fuse/nodefs"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// FIOTemplatefiles is a Filesystem implementing the FIOPlugin interface that
// first reads in a certain templatefile and then parses through all variables
// trying to call the store with the requesting users UID. If the requesting user
// does have permission for each secret, the template will be rendered with those
// secret values and returned upon an easy read syscall:
//  cat <mountpoint>/templatefiles/templated.conf
type FIOTemplatefiles struct {
	templpath string
}

// secret will be used to call the stores implementation of all the needed FUSE-
// operations together with the provided flags and fuse.Context.
type secret struct {
	flags   uint32
	context *fuse.Context
	//t       *FIOTemplatefiles
}

// GetAttr implements fuse.GetAttr
func (t *FIOTemplatefiles) GetAttr(name string, context *fuse.Context) (*fuse.Attr, fuse.Status) {
	u, err := sfsh.GetUser(context)
	if err != nil {
		return nil, fuse.EPERM
	}
	logger = sfsl.DefaultEntry(name, u)
	logger.Debug("calling operation")

	// opening directory (aka templatefiles/)
	if name == "" {
		return &fuse.Attr{
			Mode: fuse.S_IFDIR | 0550,
		}, fuse.OK
	}

	// get path to templates
	filepath := t.getCorrectPath(name)

	// check whether filepath exists
	file, err := os.Stat(filepath)
	if err != nil {
		logger.Error(err)
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
		// get filepath to templates
		filepath := t.getCorrectPath(name)
		logger.WithFields(log.Fields{"filepath": filepath}).Debug("log values")
		var flags uint32 = 0
		content, err := renderTemplatefile(filepath, flags, context)
		if err != nil {
			logger.Error(err)
			return nil, fuse.ENOENT
		}
		return &fuse.Attr{
			Mode: fuse.S_IFREG | 0550,
			Size: uint64(len(content)),
		}, fuse.OK
	}

	return nil, fuse.EINVAL
}

// OpenDir implements fuse.OpenDir
func (t *FIOTemplatefiles) OpenDir(name string, context *fuse.Context) ([]fuse.DirEntry, fuse.Status) {
	u, err := sfsh.GetUser(context)
	if err != nil {
		return nil, fuse.EPERM
	}
	logger = sfsl.DefaultEntry(name, u)
	logger.Debug("calling operation")

	// get filepath to templates
	filepath := t.getCorrectPath(name)

	// check whether filepath exists
	file, err := os.Stat(filepath)
	if err != nil {
		logger.Error(err)
		return nil, fuse.ENOENT
	}
	// check whether filepath is a directory
	// https://stackoverflow.com/questions/8824571/golang-determining-whether-file-points-to-file-or-directory
	if !file.Mode().IsDir() {
		logger.WithFields(log.Fields{"filepath": filepath}).Error("not a directory")
		return nil, fuse.ENOTDIR
	}

	entries, err := ioutil.ReadDir(filepath)
	if err != nil {
		logger.Error(err)
		return nil, fuse.EBUSY
	}
	dirs := []fuse.DirEntry{}
	for _, e := range entries {
		d := fuse.DirEntry{
			Name: e.Name(),
			Mode: uint32(e.Mode()),
		}
		dirs = append(dirs, d)
	}
	return dirs, fuse.OK
}

// Open implements fuse.Open
func (t *FIOTemplatefiles) Open(name string, flags uint32, context *fuse.Context) (nodefs.File, fuse.Status) {
	u, err := sfsh.GetUser(context)
	if err != nil {
		return nil, fuse.EPERM
	}
	logger = sfsl.DefaultEntry(name, u)
	logger.Debug("calling operation")

	// get filepath to templates
	filepath := t.getCorrectPath(name)
	logger.WithFields(log.Fields{"filepath": filepath}).Debug("log values")

	content, err := renderTemplatefile(filepath, flags, context)

	//logger.WithFields(log.Fields{"isReg": file.Mode().IsRegular()}).Debug("logging values")
	logger.Debug("returning Bytes and fuse.OK")
	logger.WithFields(log.Fields{"content": content}).Debug("log values")
	datafile := nodefs.NewDataFile(content)
	return datafile, fuse.OK
}

func renderTemplatefile(filepath string, flags uint32, context *fuse.Context) ([]byte, error) {
	// check whether filepath exists
	file, err := os.Stat(filepath)
	if err != nil {
		logger.Error(err)
		return nil, errors.New(fmt.Sprintf("Got an error for os.Stat(%s)", filepath))
	}

	// check whether filepath is a file
	// https://stackoverflow.com/questions/8824571/golang-determining-whether-file-points-to-file-or-directory
	if !file.Mode().IsRegular() {
		logger.WithFields(log.Fields{"filepath": filepath}).Error("not a directory")
		return nil, errors.New(fmt.Sprintf("%s is not a directory", filepath))
	}

	filename := path.Base(filepath)
	parser, err := template.New(filename).ParseFiles(filepath)
	// error handling
	if err != nil {
		errs := err.Error()
		logger.Error(errs)
		return nil, errors.New(fmt.Sprintf("Got an error while getting template for filepath=%s filename=%s", filepath, filename))
	}

	// https://gowalker.org/text/template#Template_Execute
	// https://yourbasic.org/golang/io-writer-interface-explained/
	// https://gowalker.org/bytes#Buffer_Bytes
	// https://stackoverflow.com/questions/23454940/getting-bytes-buffer-does-not-implement-io-writer-error-message
	var buf bytes.Buffer
	thesecret := secret{
		flags:   flags,
		context: context,
	}

	logger.WithFields(log.Fields{"err": err, "buffer": buf}).Debug("before executing parser")
	err = parser.Execute(&buf, thesecret)
	logger.WithFields(log.Fields{"err": err, "buffer": buf}).Debug("after executing parser")
	if err != nil {
		logger.Error(err)
	}
	return buf.Bytes(), err
}

// FIOPath returns name of implemented FIO plugin
func (t *FIOTemplatefiles) FIOPath() string {
	return "templatefiles"
}

// getCorrectPath returns the corrected Path for reading the file from local
// filesytem
func (t *FIOTemplatefiles) getCorrectPath(name string) string {
	return t.templpath + name
	//filepath := viper.GetString("fio.templatefiles.templatespath")+name
	//logger.WithFields(log.Fields{"filepath":filepath}).Debug("log values")
	//return filepath
}

// Get is the function that will be called from inside of the templatefile.
// You need to use following scheme to get secrets substituted:
//  {{ .Get "path/to/secret" }}
func (s secret) Get(filepath string) (string, error) {
	sto := store.GetStore()
	content, status := sto.Open(filepath, s.flags, s.context)
	logger.WithFields(log.Fields{"filepath": filepath, "content": content}).Debug("log values")
	if status != fuse.OK {
		logger.WithFields(log.Fields{"fuse.Status": status}).Error("encountered error while loading secret from store")
		//return "", errors.New("There was an error while loading Secret from store, fuse.Status="+fmt.Sprint(status))
		return "", errors.New(fmt.Sprint(status))
	}
	return content, nil
}

func init() {
	fioprov := FIOTemplatefiles{
		templpath: viper.GetString("fio.templatefiles.templatespath"),
	}
	fm := FIOMap{
		Provider: &fioprov,
	}
	RegisterProvider(&fm)
}
