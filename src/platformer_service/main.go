package main

import (
	"crypto/tls"
	"flag"
	"net"

	log "github.com/Sirupsen/logrus"
	pb "github.com/grevian/GoGame/common/platformer"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	platformer "github.com/grevian/GoGame/src/platformer_service/service"
)

func main() {
	log.Info("Server Starting up")

	var (
		listenAddr    = flag.String("listen-addr", "0.0.0.0:8077", "HTTP listen address.")
		tlsCert       = flag.String("tls-cert", "/certs/server.crt", "TLS server certificate.")
		tlsKey        = flag.String("tls-key", "/certs/server.key", "TLS server key.")
		jwtPublicKey  = flag.String("jwt-public-key", "/certs/jwt.pub.pem", "The JWT RSA public key.")
	)
	flag.Parse()

	// Load the credentials we'll use for transport security
	cert, err := tls.LoadX509KeyPair(*tlsCert, *tlsKey)
	if err != nil {
		log.WithFields(
			log.Fields{
				"tlsCert": *tlsCert,
				"tlsKey":  *tlsKey,
			}).WithError(err).Fatal("Could not load TLS Certificate")
	}

	transportCredentials := credentials.NewTLS(&tls.Config{
		InsecureSkipVerify: true,
		Certificates:       []tls.Certificate{cert},
	})

	// Create an instance of our game service
	gs, err := platformer.NewGameServer(*jwtPublicKey, transportCredentials)
	if err != nil {
		log.WithField("jwtPublicKey", *jwtPublicKey).WithError(err).Fatal("Could not start server")
	}

	s := grpc.NewServer(grpc.Creds(transportCredentials))

	// Register our game service with the grpc server
	pb.RegisterGameServerServer(s, gs)

	// Serve our grpc service on the network
	ln, err := net.Listen("tcp", *listenAddr)
	if err != nil {
		log.WithField("listenAddr", *listenAddr).WithError(err).Fatal("Failed to start listening on the network")
	}

	// This will block forever, use s.GracefulStop() or s.Stop() from a signal handler or control service or whatever
	log.Info("Starting to serve Game Service")
	err = s.Serve(ln)
	if err != nil {
		log.WithError(err).Panic("Game service terminated unexpectedly!")
	}
	log.Info("Game service terminated normally")
}
