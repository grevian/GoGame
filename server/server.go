package main

import (
	"crypto/rsa"
	"crypto/tls"
	"flag"
	"fmt"
	"io/ioutil"

	"net"

	"context"

	log "github.com/Sirupsen/logrus"
	"github.com/dgrijalva/jwt-go"
	pb "github.com/grevian/GoGame/common/platformer"
	"github.com/grevian/GoGame/server/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
)

type GameServer struct {
	jwtPublicKey *rsa.PublicKey
}

func (g *GameServer) validateTokenFromContext(ctx context.Context) (*jwt.Token, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, fmt.Errorf("expected metadata wasn't present on request")
	}

	tokenStr, ok := md["authorization"]
	if !ok {
		return nil, fmt.Errorf("expected metadata wasn't present on request")
	}

	token, err := jwt.Parse(tokenStr[0], func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
			log.WithField("algorithm", t.Header["alg"]).Error("Unexpected signing method")
			return nil, fmt.Errorf("invalid token")
		}
		return g.jwtPublicKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token string: %s", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("token was not valid: %s", token.Valid)
	}

	return token, nil
}

func NewServer(rsaPublicKey string) (*GameServer, error) {
	data, err := ioutil.ReadFile(rsaPublicKey)
	if err != nil {
		return nil, fmt.Errorf("error reading the jwt public key: %v", err)
	}

	publickey, err := jwt.ParseRSAPublicKeyFromPEM(data)
	if err != nil {
		return nil, fmt.Errorf("error parsing the jwt public key: %s", err)
	}

	return &GameServer{publickey}, nil
}

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
	authServer, err := auth.NewAuthServer(jwtPrivateKey, transportCredentials)
	if err != nil {
		log.WithError(err).Fatal("Failed to create Authorization Server")
	}

	// Serve our auth service on the network
	aln, err := net.Listen("tcp", *authAddr)
	if err != nil {
		log.WithField("authAddr", *authAddr).WithError(err).Fatal("Failed to start listening on the network")
	}
	go func() {
		err := authServer.Serve(aln)
		if err != nil {
			log.WithError(err).Error("Auth Service stopped unexpectedly")
		}
	}()

	// Create a new grpc https server that will validate against the provided transport credentials
	s := grpc.NewServer(grpc.Creds(transportCredentials))

	// Create an instance of our game service
	gs, err := NewServer(*jwtPublicKey)
	if err != nil {
		log.WithField("jwtPublicKey", *jwtPublicKey).WithError(err).Fatal("Could not start server")
	}

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
