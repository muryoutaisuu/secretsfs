// https://www.ardanlabs.com/blog/2013/11/using-log-package-in-go.html
package sfslog

import (
	"log"
	"os"
	"io"
	//"io/ioutil"
	// TODO: make debug configurable
)

type Log struct {
	Debug   *log.Logger
	Info    *log.Logger
	Warn    *log.Logger
	Error   *log.Logger
}

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
