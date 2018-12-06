// SecretsFS - Access Your Secrets Comfortably and Safely 

package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/hanwen/go-fuse/fuse/nodefs"
	"github.com/hanwen/go-fuse/fuse/pathfs"

	"github.com/Muryoutaisuu/secretsfs/cmd/secretsfs/config"
	"github.com/Muryoutaisuu/secretsfs/pkg/fio"
	"github.com/Muryoutaisuu/secretsfs/pkg/store"
	"github.com/Muryoutaisuu/secretsfs/pkg/secretsfs"
)

func main() {
	// parse arguments & flags
	flag.Usage = usage
	var defaults = flag.Bool("print-defaults", false, "prints default configurations")
	var stores = flag.Bool("print-stores", false, "prints available stores")
	var currentstore = flag.Bool("print-store", false, "prints currently set store")
	flag.Parse()

	// print default configs, -print-defaults
	if *defaults {
		fmt.Printf("Default Configs: \n%s",config.GetStringConfigDefaults())
		os.Exit(0)
	}

	// print avvailable stores, -print-stores
	if *stores {
		fmt.Printf("Available Stores are: %v\n", store.GetStores())
		os.Exit(0)
	}

	// print currently set store
	if *currentstore {
		fmt.Printf("Currently set store is: %s\n", store.GetStore().String())
		os.Exit(0)
	}

	// print usage if no arguments were provided
	if flag.NArg() != 1 {
		usage()
		os.Exit(2)
	}
	mountpoint := flag.Arg(0)

	// create the filesystem object
	sfs,err := secretsfs.NewSecretsFS(pathfs.NewDefaultFileSystem(), fio.FIOMaps(), store.GetStore())
	if err != nil {
		log.Fatal(err)
	}
	pathnfs := pathfs.NewPathNodeFs(sfs, nil)

	// create the server for the filesytem, that will mount it
	server, _, err := nodefs.MountRoot(mountpoint, pathnfs.Root(), nodefs.NewOptions())
	if err != nil {
		log.Fatalf("Mount fail: %v\n", err)
	}

	// mount the filesystem object
	server.Serve()
}

// print usage of this tool
func usage() {
	fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "  %s MOUNTPOINT\n", os.Args[0])
	flag.PrintDefaults()
}
