package main

import (
	"log"

	"github.com/georgepsarakis/periscope/service"
)

func main() {
	server, cleanup, onErr := service.NewHTTPService(service.Options{})
	defer func() {
		if err := cleanup(); err != nil {
			log.Fatal(err)
		}
	}()
	if err := server.Run(); err != nil {
		onErr(err)
	}
}
