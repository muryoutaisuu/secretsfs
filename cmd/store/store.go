package store

import (
	"log"
	"log/syslog"
)

// stores contains all registered Stores
var stores []Store

// Log is used across store package for logging needs
// uses the syslog logger, set up during init()
var Linf, Lerr *log.Logger

func init() {
	Linf,_ := syslog.NewLogger(syslog.LOG_INFO, log.LstdFlags)
	Linf.Println("Setup logging")
	//Lerr,_ := syslog.NewLogger(syslog.LOG_ERR, log.LstdFlags)
}

// RegisterStore registers Stores
func RegisterStore(s Store) {
	Linf.Println(s.String())
  stores = append(stores, s)
}

