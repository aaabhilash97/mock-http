package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/aaabhilash97/mock-http/lib/server"
)

func main() {
	userHome, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}
	defaultDefLoc := fmt.Sprintf("%s/.mock-http/definitions", userHome)

	definitions := flag.String("definitions", defaultDefLoc, "Mock definitions location")
	address := flag.String("address", "127.0.0.1:3000", "Address  Ex: 3000, 0.0.0.0:3000")
	debug := flag.Bool("debug", false, "Show debug info")
	flag.Parse()

	err = os.MkdirAll(defaultDefLoc, os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}

	err = server.StartServer(server.Options{
		Address:             *address,
		DefinitionsLocation: *definitions,
		Debug:               *debug,
	})
	if err != nil {
		log.Fatal(err)
	}
}
