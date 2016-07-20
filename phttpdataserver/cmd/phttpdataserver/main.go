package main

import (
	"flag"
	"log"
	"os"

	"github.com/mrd0ll4r/pyme-http-dataserver/phttpdataserver"
)

var (
	port    int
	workDir string
)

func init() {
	wd, err := os.Getwd()
	if err != nil {
		log.Println("Unable to get working directory")
	}
	flag.IntVar(&port, "p", 8000, "port to listen on")
	flag.StringVar(&workDir, "wd", wd, "working directory (to put files)")
}

func main() {
	flag.Parse()
	if workDir == "" {
		workDir = "."
	}

	phds := phttpdataserver.New(port, workDir)
	log.Fatal(phds.Run())
}
