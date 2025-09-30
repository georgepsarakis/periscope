package main

import (
	"github.com/georgepsarakis/periscope/service"
)

func main() {
	server, cleanup, onErr := service.NewHTTPService(service.Options{})
	defer cleanup()
	if err := server.Run(); err != nil {
		onErr(err)
	}
}
