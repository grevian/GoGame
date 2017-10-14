package client

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"

	log "github.com/Sirupsen/logrus"
	pb_auth "github.com/grevian/GoGame/common/auth"
	pb "github.com/grevian/GoGame/common/platformer"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

const GAMESERVER_ADDRESS = "localhost:8078"

type NetworkClient struct {
	tlsCredentials credentials.TransportCredentials
	rpcCredentials credentials.PerRPCCredentials
	gameClient     pb.GameServerClient
}

func NewNetworkClient(username *string, password *string, certPath *string, jwtPublicKeyPath *string) (*NetworkClient, error) {
	log.Info("Connecting to Network")

	// Load our CA Information for transport security
	rawCACert, err := ioutil.ReadFile(*certPath)
	if err != nil {
		log.WithField("certPath", *certPath).WithError(err).Error("Could not read CA")
		return nil, err
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(rawCACert)

	transportCredentials := credentials.NewTLS(&tls.Config{
		RootCAs: caCertPool,
	})

	// Instantiate a client for the auth service, that will fetch per-request credentials
	authTokenFetcher, err := NewAuthServiceTokenFetcher(&pb_auth.Credentials{Username: *username, Password: *password}, transportCredentials, jwtPublicKeyPath)
	if err != nil {
		log.WithError(err).Error("Could not construct auth service token fetcher")
		return nil, err
	}

	// Dial the game server with all our credentials in place
	conn, err := grpc.Dial(GAMESERVER_ADDRESS,
		grpc.WithTransportCredentials(transportCredentials),
		grpc.WithPerRPCCredentials(authTokenFetcher),
	)

	if err != nil {
		log.WithField("GAMESERVER_ADDRESS", GAMESERVER_ADDRESS).WithError(err).Error("Could not access game server")
		return nil, err
	}

	// Connect to the game service
	gameClient := pb.NewGameServerClient(conn)

	return &NetworkClient{
		tlsCredentials: transportCredentials,
		rpcCredentials: authTokenFetcher,
		gameClient:     gameClient,
	}, nil
}

func (l *NetworkClient) Update(character *Character) {
	// Send an update on character positions
}
