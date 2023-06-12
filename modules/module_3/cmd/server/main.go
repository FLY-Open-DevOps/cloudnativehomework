package main

import (
	"log"
	"module3/server"
)

const addr = "0.0.0.0:8080"

func main() {
	if err := server.Serve(addr); err != nil {
		log.Fatalf("server stop because: %v", err)
	}
}
