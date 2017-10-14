package main

import (
	"crypto/tls"
	"flag"

	"net"

	log "github.com/Sirupsen/logrus"
	pb "github.com/grevian/GoGame/common/platformer"
	"github.com/grevian/GoGame/server/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"./platformer"
)

func main() {
	log.Info("Server Starting up")

	var (
		listenAddr    = flag.String("listen-addr", "0.0.0.0:8077", "HTTP listen address.")
		authAddr      = flag.String("auth-listen-addr", "0.0.0.0:8078", "HTTP listen address.")
		tlsCert       = flag.String("tls-cert", "/cert.pem", "TLS server certificate.")
		tlsKey        = flag.String("tls-key", "/key.pem", "TLS server key.")
		jwtPublicKey  = flag.String("jwt-public-key", "/jwt.pem", "The JWT RSA public key.")
		jwtPrivateKey = flag.String("jwt-private-key", "/jwt-private.pem", "The JWT RSA private key.")
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
		Certificates: []tls.Certificate{cert},
	})

	// Create our Authorization server
	authService, err := auth.NewAuthServer(jwtPrivateKey, transportCredentials)

	// Serve our auth service on the network
	aln, err := net.Listen("tcp", *authAddr)
	if err != nil {
		log.WithField("authAddr", *authAddr).WithError(err).Fatal("Failed to start listening on the network")
	}
	go func() {
		err := authService.Serve(aln)
		if err != nil {
			log.WithError(err).Error("Auth Service stopped unexpectedly")
		}
	}()

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
	err = s.Serve(ln)
	if err != nil {
		log.WithError(err).Panic("Server terminated unexpectedly!")
	}
	log.Info("Server terminated normally")
}
