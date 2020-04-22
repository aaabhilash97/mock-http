package main

import (
	"flag"

	"log"

	"github.com/aaabhilash97/mock-http/lib/server"
)

func main() {
	definitions := flag.String("definitions", "", "Mock definitions location")
	address := flag.String("address", "", "Address")
	flag.Parse()

	err := server.StartServer(server.Options{
		Address:             *address,
		DefinitionsLocation: *definitions,
	})
	if err != nil {
		log.Fatal(err)
	}
}
