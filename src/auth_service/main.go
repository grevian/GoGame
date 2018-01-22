package main

import (
	"flag"
	"net"
	"os"
	"os/signal"

	log "github.com/Sirupsen/logrus"
	pb "github.com/grevian/GoGame/common/auth"
	"google.golang.org/grpc"

	authService "github.com/grevian/GoGame/src/auth_service/service"
)

func main() {
	var (
		listenAddr        = flag.String("listen-addr", "0.0.0.0:8076", "gRPC Authorization server address to listen on")
		tlsCert           = flag.String("tls-cert", "/certs/server.crt", "TLS server certificate.")
		tlsKey            = flag.String("tls-key", "/certs/server.key", "TLS server key.")
		jwtPrivateKeyPath = flag.String("jwt-private-key", "/certs/jwt.key", "The JWT RSA private key.")
	)
	flag.Parse()

	// Load secrets and credentials
	transportCredentials := loadTransportCredentials(tlsCert, tlsKey)
	jwtPrivateKey := loadJWTPrivateKey(jwtPrivateKeyPath)

	// Create our Authorization server
	s := grpc.NewServer(grpc.Creds(transportCredentials))
	authService, err := authService.NewAuthServer(jwtPrivateKey)
	pb.RegisterAuthServerServer(s, authService)

	// Register a signal handler to gently shut down the service
	outerSignalChannel := make(chan os.Signal, 1)
	signal.Notify(outerSignalChannel, os.Interrupt)
	signal.Notify(outerSignalChannel, os.Kill)

	go func(signalChannel <-chan os.Signal) {
		for {
			// Block until a signal is received.
			signal := <-signalChannel
			if signal == os.Kill || signal == os.Interrupt {
				log.Info("Received Kill signal, shutting down Authorization service")
				// Interrupt the current RPC streams and release any other resources
				authService.GracefulStop()

				// Stop new RPCs from being accepted, blocks until all RPCs are terminated
				s.GracefulStop()
				log.Info("Successfully shut down Authorization service")
			} else {
				log.WithField("signal", signal).Warning("Unexpected signal encountered")
			}
		}
	}(outerSignalChannel)

	listener, err := net.Listen("tcp", *listenAddr)
	if err != nil {
		log.WithField("authAddr", *listenAddr).WithError(err).Fatal("Failed to start listening on the network")
	}

	log.Info("Starting to serve Authorization service")
	// This will block until the signal handler shuts down the service, or an error occurs
	err = s.Serve(listener)
	if err != nil {
		if err == grpc.ErrServerStopped {
			log.WithError(err).Info("Serving stopped because the service was asked to stop")
		} else {
			log.WithError(err).Error("Authorization service stopped unexpectedly")
		}
	}
	log.Info("Authorization service stopped")
}
