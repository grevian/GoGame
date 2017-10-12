package main

import (
	log "github.com/Sirupsen/logrus"

	"./client"
)

func main() {
	log.Print("Starting up")
	client.NewClient()
}
