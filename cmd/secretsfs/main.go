// SecretsFS - Access Your Secrets Comfortably and Safely

package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	"github.com/muryoutaisuu/secretsfs/cmd/secretsfs/config"
	sfs "github.com/muryoutaisuu/secretsfs/pkg/secretsfs"
	"github.com/muryoutaisuu/secretsfs/pkg/store"
)

var (
	// Build Time injection
	Version   string = ""
	BuildDate string = ""
)

func main() {
	// ARGUMENT THINGIES START
	// parse flags and arguments
	flag.Usage = usage
	var opts = flag.String("o", "", "Mount options passed through to fuse")
	var currentstore = flag.Bool("print-store", false, "prints currently set store")
	var defaults = flag.Bool("print-defaults", false, "prints default configurations")
	var stores = flag.Bool("print-stores", false, "prints available stores")
	var fios = flag.Bool("print-fios", false, "prints available FIOs")
	var json = flag.Bool("log-json", false, "log in json format")
	var fusedebug = flag.Bool("fuse-debug", false, "debug logging of fuse library")
	var printversion = flag.Bool("version", false, "print version information")

	flag.CommandLine.Parse(os.Args[firstDashedArg():])

	if *printversion {
		fmt.Printf("secretsfs version: %v\n", Version)
		fmt.Printf("build date       : %v\n", BuildDate)
		os.Exit(0)
	}

	// print default configs, --print-defaults
	if *defaults {
		fmt.Printf(config.GetStringConfigDefaults())
		os.Exit(0)
	}

	// print available stores, --print-stores
	if *stores {
		fmt.Printf("Available Stores are: %v\n", store.GetStores())
		os.Exit(0)
	}

	// prints available fios, --print-fios
	if *fios {
		list := make([]string, 0)
		for k := range sfs.FIOMaps() {
			list = append(list, k)
		}
		fmt.Printf("Available FIOs are: %v\n", list)
		os.Exit(0)
	}

	// print currently set store
	if *currentstore {
		fmt.Printf("Currently set store is: %s\n", (*store.GetStore()).String())
		os.Exit(0)
	}

	// setup logging
	//log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	log.SetOutput(os.Stdout)
	log.SetReportCaller(true)
	if *json {
		log.SetFormatter(&log.JSONFormatter{})
	} else {
		log.SetFormatter(&log.TextFormatter{FullTimestamp: true, DisableColors: true})
	}
	l, err := log.ParseLevel(viper.GetString("general.logging.level"))
	if err != nil {
		log.Error("Could not parse logging Level configuration! Will fallback to info level")
		log.SetLevel(log.InfoLevel)
	} else {
		log.WithFields(log.Fields{"newloglevel": l}).Info("setting new loglevel")
		log.SetLevel(l)
	}

	// print usage if no arguments were provided
	// os.Args[0] is the programname itself
	if len(os.Args) < 2 {
		log.WithFields(log.Fields{"os.Args": os.Args}).Error("not enough arguments, showing usage")
		usage()
		os.Exit(1)
	}

	log.Info("log values")
	mountpoint := os.Args[1]
	if fileinfo, err := os.Stat(mountpoint); err != nil {
		log.WithFields(log.Fields{"mountpoint": mountpoint, "error": err}).Error("Got error while doing os.Stat for mountpoint")
		os.Exit(2)
	} else if !fileinfo.IsDir() {
		log.WithFields(log.Fields{"mountpoint": mountpoint}).Error("can't mount, mountpoint is not a directory")
		os.Exit(2)
	}
	log.WithFields(log.Fields{"mountpoint": mountpoint}).Debug("log values")
	// ARGUMENT THINGIES END

	// This is where we'll mount the FS
	fms := sfs.FIOMapsEnabled()
	log.Debugf("fms: %v\n", fms)
	root := sfs.GetNewRootNode("/", fms)
	// options
	fsopts := fuse.MountOptions{}
	fsopts.Options = strings.Split(*opts, ",")
	if *fusedebug {
		fsopts.Debug = true
	}
	server, err := fs.Mount(mountpoint, root, &fs.Options{
		MountOptions: fsopts,
	})
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Errorf("error while mounting %s", os.Args[0])
		os.Exit(3)
	}

	log.WithFields(log.Fields{"mountpoint": mountpoint}).Infof("%s mounted", os.Args[0])
	log.Infof("Unmount by calling 'fusermount -u %s'", mountpoint)

	// Wait until unmount before exiting
	log.Infof("Serving now...")
	server.Wait()
}

func usage() {
	fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "  %s [OPTIONS]\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "  %s [MOUNTPOINT] [OPTIONS]\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "OPTIONS:\n")
	flag.PrintDefaults()
}

// firstDashedArg returns the index of the first dashed argument, e.g. -ex
// https://stackoverflow.com/a/51526473/4069534
func firstDashedArg() int {
	for k, v := range os.Args {
		if len(v) > 0 && v[0] == '-' {
			return k
		}
	}
	return 1
}
