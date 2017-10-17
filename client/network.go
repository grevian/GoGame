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
	"context"
)

const GAMESERVER_ADDRESS = "localhost:8077"

type NetworkClient struct {
	tlsCredentials credentials.TransportCredentials
	rpcCredentials credentials.PerRPCCredentials
	gameClient     pb.GameServerClient

	// Bidirectional streams used to send and receive server interactions
	commandStream pb.GameServer_CommandUpdatesClient
	positionStream pb.GameServer_PositionUpdatesClient
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
		InsecureSkipVerify: true,
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

	commandStream, err := gameClient.CommandUpdates(context.Background())
	if err != nil {
		log.WithError(err).Error("Could not open command stream")
		return nil, err
	}

	positionStream, err := gameClient.PositionUpdates(context.Background())
	if err != nil {
		log.WithError(err).Error("Could not open position stream")
		return nil, err
	}

	// Load position updates waiting from any other characters
	go func() {
		for {
			positionUpdate, err := positionStream.Recv()
			if err != nil {
				log.WithError(err).Error("Unexpected error on player position data stream")
				return
			}
			// TODO Apply position update
			// TODO Need to change position stream to deal with wrapped positions that include player data
			log.WithField("User", positionUpdate.UserIdentifier).Debug("Position Updated")
		}
	}()


	return &NetworkClient{
		tlsCredentials: transportCredentials,
		rpcCredentials: authTokenFetcher,
		gameClient:     gameClient,

		commandStream: commandStream,
		positionStream: positionStream,
	}, nil
}

func (l *NetworkClient) Update(character *Character) {
	// Send an update on character positions
	position := pb.Position{
		X: 45.0,
		Y: 45.0,
	}

	if l.positionStream != nil {
		err := l.positionStream.Send(&position)

		if err != nil {
			log.WithError(err).Error("Failed to send position update")
			l.positionStream.CloseSend()
			l.positionStream = nil
		}
	}
}
