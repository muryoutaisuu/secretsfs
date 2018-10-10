package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"github.com/Muryoutaisuu/secretsfs/pkg/secretsfs"
	//"sync/atomic"
	//"syscall"
	//"time"

	//"bazil.org/fuse"
	//"bazil.org/fuse/fs"
	//_ "bazil.org/fuse/fs/fstestutil"
	//"bazil.org/fuse/fuseutil"
	//"golang.org/x/net/context"

	//"github.com/Muryoutaisuu/secretsfs/cmd/store"
)

func main() {
  flag.Usage = usage
		flag.Parse()

	if flag.NArg() != 1 {
		usage()
		os.Exit(2)
	}
	mountpoint := flag.Arg(0)

	if err := secretsfs.Run(mountpoint); err != nil {
		log.Fatal(err)
	}
}

func usage() {
	fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "  %s MOUNTPOINT\n", os.Args[0])
	flag.PrintDefaults()
}
