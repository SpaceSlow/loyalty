package main

import (
	"github.com/SpaceSlow/loyalty/internal/server"
	"log"
)

func main() {
	if err := server.RunServer(); err != nil {
		log.Fatalf("Error occured: %s.\r\nExiting...", err)
	}
}
