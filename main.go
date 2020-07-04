package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/aaabhilash97/mock-http/lib/server"
)

var defaultDefLoc string
var definitions string
var address string
var debug bool

func init() {
	userHome, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}
	defaultDefLoc = fmt.Sprintf("%s/.mock-http/definitions", userHome)

	flag.StringVar(&definitions, "definitions", defaultDefLoc, "Mock definitions location")
	flag.StringVar(&address, "address", "127.0.0.1:3000", "Address  Ex: 3000, 0.0.0.0:3000")
	flag.BoolVar(&debug, "debug", false, "Show debug info")
	flag.Parse()
}

func main() {
	err := os.MkdirAll(defaultDefLoc, os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}

	err = server.StartServer(server.Options{
		Address:             address,
		DefinitionsLocation: definitions,
		Debug:               debug,
	})
	if err != nil {
		log.Fatal(err)
	}
}
