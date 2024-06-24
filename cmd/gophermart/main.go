package main

import (
	"log"

	"github.com/SpaceSlow/loyalty/internal/server"
)

func main() {
	if err := server.RunServer(); err != nil {
		log.Fatalf("Error occured: %s.\r\nExiting...", err)
	}
}
