package main

import (
	"flag"

	log "github.com/Sirupsen/logrus"

	//	"./client"
	"github.com/grevian/GoGame/client"
)

func main() {
	log.Print("Starting up")
	var serverAddr = flag.String("server-addr", "gogame.grevian.org", "gameserver address")
	flag.Parse()
	client.NewClient(serverAddr)
}
