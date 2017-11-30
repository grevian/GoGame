package main

import (
	log "github.com/Sirupsen/logrus"

	"./client"
)

func main() {
	log.Print("Starting up")
	var serverAddr = flag.String("server-addr", "gogame.grevian.org:8077", "gameserver address")
	flag.Parse()
	client.NewClient(serverAddr)
}
