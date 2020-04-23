package main

import (
	"flag"
	"fmt"
	"os"

	"log"

	"github.com/aaabhilash97/mock-http/lib/server"
)

func main() {

	userHome, _ := os.UserHomeDir()
	defaultDefLoc := fmt.Sprintf("%s/.mock-http/definitions", userHome)
	_ = os.MkdirAll(defaultDefLoc, os.ModePerm)

	definitions := flag.String("definitions", defaultDefLoc, "Mock definitions location")
	address := flag.String("address", "127.0.0.1:3000", "Address  Ex: 3000, 0.0.0.0:3000")
	flag.Parse()

	err := server.StartServer(server.Options{
		Address:             *address,
		DefinitionsLocation: *definitions,
	})
	if err != nil {
		log.Fatal(err)
	}
}
