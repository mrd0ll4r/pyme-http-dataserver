package main

import (
	"flag"
	"log"
	"os"

	"github.com/mrd0ll4r/pyme-http-dataserver/phttpdataserver"
)

var (
	port     int
	workDir  string
	doLog    bool
	testMode bool
)

func init() {
	wd, err := os.Getwd()
	if err != nil {
		log.Println("Unable to get working directory")
	}
	flag.IntVar(&port, "p", 8000, "port to listen on")
	flag.StringVar(&workDir, "wd", wd, "working directory (to put files)")
	flag.BoolVar(&doLog, "l", true, "write request logs")
	flag.BoolVar(&testMode, "t", false, "run in test mode, i.e. discard all data")
}

func main() {
	flag.Parse()
	if workDir == "" {
		workDir = "."
	}

	s := phttpdataserver.New(port, workDir, doLog, testMode)
	log.Fatal(s.Run())
}
