package secretsfs

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"syscall"
	"text/template"

	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	"github.com/muryoutaisuu/secretsfs/pkg/store"
)

var TEMPLATESPATHS map[string]string

// secret will be used to call the stores implementation of all the needed FUSE-
// operations together with the provided flags and fuse.Context.
type secret struct {
	ctx *context.Context
}

// Get is the function that will be called from inside of the templatefile.
// You need to use following scheme to get secrets substituted:
//  {{ .Get "path/to/secret" }}
func (s secret) Get(filepath string) (string, error) {
	sto := *store.GetStore()
	sec, err := sto.GetSecret(filepath, *s.ctx)
	if err != nil {
		return "", err
	}
	if sec.Content == "" {
		return "", fmt.Errorf("msg=\"content of secret is empty\" secret=\"%v\"\n", filepath)
	}
	return sec.Content, nil
}

type FIOTemplateFiles struct{}

var _ = (FIORoot)((*FIOTemplateFiles)(nil))

func (sf *FIOTemplateFiles) Readdir(n *SfsNode, ctx context.Context) (out fs.DirStream, errno syscall.Errno) {
	log.WithFields(log.Fields{
		"n":                   n,
		"n.npath":             n.npath,
		"IsRootPath(n.npath)": IsRootPath(n.npath)}).Debug("log values")

	var direntries []fuse.DirEntry
	rtemplp, utemplp := getTemplateSubPaths(n.npath) // roottemplatepath + unixtemplatepath
	// return root template paths
	if IsRootPath(n.npath) {
		for k := range TEMPLATESPATHS {
			fixedpath := sf.prefixPath(k)
			direntries = append(direntries, fuse.DirEntry{
				Name: filepath.Base(fixedpath),
				Ino:  GetInode(fixedpath),
				Mode: fuse.S_IFDIR,
			})
		}

		// walk unixpaths and return their dir listings
	} else if templp, ok := TEMPLATESPATHS[rtemplp]; ok {
		unixpath := filepath.Join(templp, utemplp)
		files, err := ioutil.ReadDir(unixpath)
		if err != nil {
			log.WithFields(log.Fields{"unixpath": unixpath, "templp": templp, "utemplp": utemplp, "error": err}).Error("got error while reading dir contents of templatepath")
			return nil, syscall.ENOENT
		}
		for _, f := range files {
			direntries = append(direntries, fuse.DirEntry{
				Name: f.Name(),
				Ino:  GetInode(filepath.Join(n.npath, f.Name())),
				Mode: getModeFromFileInfo(f),
			})
		}
	} else {
		return nil, syscall.ENOSYS
	}

	log.WithFields(log.Fields{"direntries": direntries}).Debug("log values")
	return fs.NewListDirStream(direntries), fs.OK
}

func (sf *FIOTemplateFiles) Lookup(n *SfsNode, ctx context.Context, name string, out *fuse.EntryOut) (node *fs.Inode, errno syscall.Errno) {
	log.WithFields(log.Fields{
		"n":          n,
		"n.npath":    n.npath,
		"name":       name,
		"out.NodeId": out.NodeId}).Debug("log values")

	prefixedfullname := filepath.Join(n.npath, name)
	// if is root template path, then
	if _, ok := TEMPLATESPATHS[name]; ok {
		return getLookupChild(n, prefixedfullname, fuse.S_IFDIR, ctx, out)
	}

	// walk unixpaths and return their dir listings
	rtemplp, utemplp := getTemplateSubPaths(n.npath) // roottemplatepath + unixtemplatepath
	if templp, ok := TEMPLATESPATHS[rtemplp]; ok {
		unixpath := filepath.Join(templp, utemplp)
		files, err := ioutil.ReadDir(unixpath)
		if err != nil {
			log.WithFields(log.Fields{"unixpath": unixpath, "templp": templp, "utemplp": utemplp, "error": err}).Error("got error while reading dir contents of templatepath")
			return nil, syscall.ENOENT
		}
		for _, f := range files {
			// if upath listing contains the requested filename
			if f.Name() == name {
				return getLookupChild(n, prefixedfullname, getModeFromFileInfo(f), ctx, out)
			}
		}
	}
	return nil, syscall.ENOENT
}

func (sf *FIOTemplateFiles) Open(n *SfsNode, ctx context.Context, flags uint32) (fh fs.FileHandle, fuseFlags uint32, errno syscall.Errno) {
	return nil, 0, 0
}

func (sf *FIOTemplateFiles) Read(n *SfsNode, ctx context.Context, f fs.FileHandle, dest []byte, off int64) (fuse.ReadResult, syscall.Errno) {
	log.WithFields(log.Fields{"n": n, "n.npath": n.npath}).Debug("log values")

	rtemplp, utemplp := getTemplateSubPaths(n.npath) // roottemplatepath + unixtemplatepath
	if templp, ok := TEMPLATESPATHS[rtemplp]; ok {
		unixpath := filepath.Join(templp, utemplp)
		log.WithFields(log.Fields{
			"rtemplp":  rtemplp,
			"utemplp":  utemplp,
			"templp":   templp,
			"unixpath": unixpath}).Debug("log values")
		content, err := renderTemplatefile(unixpath, &ctx)
		if err != nil {
			log.WithFields(log.Fields{
				"rtemplp":  rtemplp,
				"utemplp":  utemplp,
				"templp":   templp,
				"unixpath": unixpath,
				"error":    err}).Error("got error while rendering templatefile")
			return nil, syscall.EIO
		}
		results := fuse.ReadResultData([]byte(content))
		return results, fs.OK
	}
	return nil, syscall.ENOENT
}

func (sf *FIOTemplateFiles) Getattr(n *SfsNode, ctx context.Context, fh fs.FileHandle, out *fuse.AttrOut) syscall.Errno {
	log.WithFields(log.Fields{
		"n":                   n,
		"n.npath":             n.npath,
		"IsRootPath(n.npath)": IsRootPath(n.npath)}).Debug("log values")

	// if rootpath
	if IsRootPath(n.npath) {
		out.Ino = GetInode(n.npath)
		return fs.OK
	}

	rtemplp, utemplp := getTemplateSubPaths(n.npath) // roottemplatepath + unixtemplatepath
	// if is root template path, then
	if _, ok := TEMPLATESPATHS[rtemplp]; ok && utemplp == "" {
		out.Ino = GetInode(n.npath)
		return fs.OK
	}

	// walk unixpath and lstat on requested file
	log.WithFields(log.Fields{
		"rtemplp":                 rtemplp,
		"utemplp":                 utemplp,
		"TEMPLATESPATHS[rtemplp]": TEMPLATESPATHS[rtemplp]}).Debug("log values")
	if templp, ok := TEMPLATESPATHS[rtemplp]; ok {
		unixpath := filepath.Join(templp, utemplp)
		log.Printf("unixpath=\"%v\"\n", unixpath)
		log.WithFields(log.Fields{
			"rtemplp":                 rtemplp,
			"utemplp":                 utemplp,
			"TEMPLATESPATHS[rtemplp]": TEMPLATESPATHS[rtemplp],
			"unixpath":                unixpath}).Debug("log values")
		fileinfo, err := os.Stat(unixpath)
		if err != nil {
			log.WithFields(log.Fields{
				"rtemplp":                 rtemplp,
				"utemplp":                 utemplp,
				"TEMPLATESPATHS[rtemplp]": TEMPLATESPATHS[rtemplp],
				"unixpath":                unixpath,
				"error":                   err}).Error("got error while performing os.Stat(unixpath)")
			return syscall.ENOENT
		}
		if fileinfo.Mode().IsRegular() {
			content, err := renderTemplatefile(unixpath, &ctx)
			if err != nil {
				log.WithFields(log.Fields{"error": err}).Error("error while rendering templatefile for size calculation")
			}
			out.Size = uint64(len(content))
		}
		out.Ino = GetInode(n.npath)
		return fs.OK
	}
	return syscall.ENOENT
}

func (sf *FIOTemplateFiles) FIOPath() string {
	return "templatefiles"
}

func (sf *FIOTemplateFiles) prefixPath(npath string) string {
	return string(filepath.Separator) + filepath.Join(sf.FIOPath(), npath)
}

func getTemplateSubPaths(npath string) (rtemplp, utemplp string) {
	_, spath := rootName(npath) // rpath + spath   == rootpath + subpath
	return rootName(spath)      // roottemplatepath + unixtemplatepath
}

func getLookupChild(n *SfsNode, name string, mode uint32, ctx context.Context, out *fuse.EntryOut) (child *fs.Inode, errno syscall.Errno) {
	ino := GetInode(name)
	stable := fs.StableAttr{
		Mode: mode,
		Ino:  ino,
	}
	operations := NewNode(name)
	child = n.NewInode(ctx, operations, stable)
	out.NodeId = ino
	return child, fs.OK
}

// tpath = templatepath
func renderTemplatefile(tpath string, context *context.Context) ([]byte, error) {
	// check whether filepath exists
	fileinfo, err := os.Stat(tpath)
	if err != nil {
		return nil, fmt.Errorf(fmt.Sprintf("msg=\"got an error for os.Stat(%s)\" error=\"%v\"\n", tpath, err))
	}

	// check whether filepath is a file
	// https://stackoverflow.com/questions/8824571/golang-determining-whether-file-points-to-file-or-directory
	if !fileinfo.Mode().IsRegular() {
		log.WithFields(log.Fields{"tpath": tpath}).Error("file is not a regular file, can not render templatefile")
		return nil, fmt.Errorf(fmt.Sprintf("%s is not a file", tpath))
	}

	filename := filepath.Base(tpath)
	parser, err := template.New(filename).ParseFiles(tpath)
	// error handling
	if err != nil {
		return nil, fmt.Errorf("msg=\"Got an error while getting template\" filepath=\"%s\" filename=\"%s\" error=\"%v\"\n", tpath, filename, err)
	}

	// https://gowalker.org/text/template#Template_Execute
	// https://yourbasic.org/golang/io-writer-interface-explained/
	// https://gowalker.org/bytes#Buffer_Bytes
	// https://stackoverflow.com/questions/23454940/getting-bytes-buffer-does-not-implement-io-writer-error-message
	var buf bytes.Buffer
	thesecret := secret{
		ctx: context,
	}

	err = parser.Execute(&buf, thesecret)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), err
}

func generateTemplatesPaths() {
	TEMPLATESPATHS = viper.GetStringMapString("fio.templatefiles.templatespaths")
}

func init() {
	fioroot := FIOTemplateFiles{}
	fm := FIOMap{
		Root: &fioroot,
	}
	RegisterRoot(&fm)
	generateTemplatesPaths()
}
