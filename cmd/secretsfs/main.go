// SecretsFS - Access Your Secrets Comfortably and Safely 

package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/hanwen/go-fuse/fuse"
	"github.com/hanwen/go-fuse/fuse/nodefs"
	"github.com/hanwen/go-fuse/fuse/pathfs"

	"github.com/muryoutaisuu/secretsfs/cmd/secretsfs/config"
	"github.com/muryoutaisuu/secretsfs/pkg/fio"
	"github.com/muryoutaisuu/secretsfs/pkg/store"
	"github.com/muryoutaisuu/secretsfs/pkg/secretsfs"
)

func main() {
	// ARGUMENT THINGIES START
	// parse arguments & flags
	flag.Usage = usage
	var opts = flag.String("o","","Options passed through to fuse")
	var currentstore = flag.Bool("print-store", false, "prints currently set store")
	var defaults = flag.Bool("print-defaults", false, "prints default configurations")
	var stores = flag.Bool("print-stores", false, "prints available stores")
	var fios = flag.Bool("print-fios", false, "prints available FIOs")

	firstdashed := firstDashedArg(os.Args)
	flag.CommandLine.Parse(os.Args[firstdashed:])

	// print default configs, -print-defaults
	if *defaults {
		fmt.Printf("Default Configs: \n%s",config.GetStringConfigDefaults())
		os.Exit(0)
	}

	// print available stores, -print-stores
	if *stores {
		fmt.Printf("Available Stores are: %v\n", store.GetStores())
		os.Exit(0)
	}

	// prints available fios, -print-fios
	if *fios {
		maps := fio.FIOMaps()
		list := make([]string, 0)
		for k := range maps {
			list = append(list, k)
		}
		fmt.Printf("Available FIOs are: %v\n", list)
		os.Exit(0)
	}

	// print currently set store
	if *currentstore {
		fmt.Printf("Currently set store is: %s\n", store.GetStore().String())
		os.Exit(0)
	}

	log.Printf("Call is: %s\n",os.Args)
	// print usage if no arguments were provided
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}
	mountpoint := os.Args[1]
	log.Println("mountpoint is: "+mountpoint)
	// ARGUMENT THINGIES END

	// create the filesystem object
	sfs,err := secretsfs.NewSecretsFS(pathfs.NewDefaultFileSystem(), fio.FIOMaps(), store.GetStore())
	if err != nil {
		log.Fatal(err)
	}
	pathnfs := pathfs.NewPathNodeFs(sfs, nil)

	fsc := nodefs.NewFileSystemConnector(pathnfs.Root(), nodefs.NewOptions())  // FileSystemConnector

	// set options
	fsopts := fuse.MountOptions{}
	log.Println(*opts)
	fsopts.Options = strings.Split(*opts, ",")

	// create server
	server, err := fuse.NewServer(fsc.RawFS(), mountpoint, &fsopts)
	if err != nil {
		log.Printf("Mountfail: %v\n", err)
		os.Exit(1)
	}
	// mount and now serve me till the end!!!
	server.Serve()
	defer server.Unmount()
	return
}

// print usage of this tool
func usage() {
	fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "  %s MOUNTPOINT\n", os.Args[0])
	flag.PrintDefaults()
}

// firstDashedArg returns the index of the first dashed argument, e.g. -ex
// https://stackoverflow.com/a/51526473/4069534
func firstDashedArg(args []string) int {
	for i := 1; i < len(args); i ++ {
		if len(args[i]) > 0 && args[i][0] == '-' {
			return i
		}
	}
	return 1
}
