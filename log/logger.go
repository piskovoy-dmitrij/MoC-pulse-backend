package log

import (
	"log"
	"os"
	"io/ioutil"
)

var Error *log.Logger
var Warning *log.Logger
var Info *log.Logger
var Debug *log.Logger
var file *os.File

func NewLogger(logLevel int64, logPath string) {
	var err error
	Error = log.New(ioutil.Discard, "ERROR: ", log.Ldate|log.Lmicroseconds|log.Lshortfile)
	Warning = log.New(ioutil.Discard, "WARNING: ", log.Ldate|log.Lmicroseconds|log.Lshortfile)
	Info = log.New(ioutil.Discard, "INFO: ", log.Ldate|log.Lmicroseconds|log.Lshortfile)
	Debug = log.New(ioutil.Discard, "DEBUG: ", log.Ldate|log.Lmicroseconds|log.Lshortfile)
	if logLevel >= 1 {
		file, err = os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			log.Fatalf("Failed to open log file %s: %v\n", logPath, err)
		}
		Error = log.New(file, "ERROR: ", log.Ldate|log.Lmicroseconds|log.Lshortfile)
		if logLevel >= 2 {
			Warning = log.New(file, "WARNING: ", log.Ldate|log.Lmicroseconds|log.Lshortfile)
			if logLevel >= 3 {
				Info = log.New(file, "INFO: ", log.Ldate|log.Lmicroseconds|log.Lshortfile)
				if logLevel >= 4 {
					Debug = log.New(file, "DEBUG: ", log.Ldate|log.Lmicroseconds|log.Lshortfile)
				}
			}
		}
	}
}

func CloseLog() {
	file.Close()
}
