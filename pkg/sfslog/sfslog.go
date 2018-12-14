// sfslog package provides all needed logging features for a consistent logging
// format through all provided plugins.
// Code taken from example:
// https://www.ardanlabs.com/blog/2013/11/using-log-package-in-go.html
package sfslog

import (
	"log"
	"os"
	"io"
	//"io/ioutil"
	// TODO: make debug configurable
)

// Log will contain four different logging levels, which themselves can be
// called like any other default go logger (because they are default go logger
// in reality).
type Log struct {
	Debug   *log.Logger
	Info    *log.Logger
	Warn    *log.Logger
	Error   *log.Logger
}

// Logger return a struct of type Log, which contains for different logging level
// default go loggers.
func Logger() *Log {
	var l Log
	// log setup
	logInit(&l, os.Stdout, os.Stdout, os.Stdout, os.Stderr)
	//logInit(&l, ioutil.Discard, os.Stdout, os.Stdout, os.Stderr)
	return &l
}

func logInit(
	l *Log,
	debugHandle io.Writer,
 	infoHandle io.Writer,
 	warnHandle io.Writer,
 	errorHandle io.Writer) {

 	l.Debug = log.New(debugHandle,
		"DEBUG: ",
		log.Ldate|log.Ltime|log.Lshortfile)

 	l.Info = log.New(infoHandle,
		"INFO: ",
		log.Ldate|log.Ltime|log.Lshortfile)

 	l.Warn = log.New(warnHandle,
		"WARN: ",
		log.Ldate|log.Ltime|log.Lshortfile)

 	l.Error = log.New(errorHandle,
		"ERROR: ",
		log.Ldate|log.Ltime|log.Lshortfile)
}
