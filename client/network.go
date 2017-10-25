package client

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"

	"context"

	log "github.com/Sirupsen/logrus"
	pb_auth "github.com/grevian/GoGame/common/auth"
	pb "github.com/grevian/GoGame/common/platformer"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

const GAMESERVER_ADDRESS = "localhost:8077"

type NetworkClient struct {
	tlsCredentials credentials.TransportCredentials
	rpcCredentials credentials.PerRPCCredentials
	gameClient     pb.GameServerClient

	// Bidirectional stream used to send and receive server interactions
	updateStream pb.GameServer_ConnectClient

	networkClock int64
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
		RootCAs:            caCertPool,
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

	updateStream, err := gameClient.Connect(context.Background())
	if err != nil {
		log.WithError(err).Error("Could not open update stream")
		return nil, err
	}

	return &NetworkClient{
		tlsCredentials: transportCredentials,
		rpcCredentials: authTokenFetcher,
		gameClient:     gameClient,

		updateStream: updateStream,

		networkClock: 0,
	}, nil
}

func (l *NetworkClient) Reset(character *Character) {
	quitCmd := pb.ClientUpdate{
		UpdatePayload: &pb.ClientUpdate_C{
			C: &pb.Command{
				Command: pb.Command_QUIT,
			},
		},
	}

	joinCmd := pb.ClientUpdate{
		UpdatePayload: &pb.ClientUpdate_C{
			C: &pb.Command{
				Command: pb.Command_JOINED,
			},
		},
	}

	l.updateStream.Send(&quitCmd)
	l.updateStream.Send(&joinCmd)
}

func (l *NetworkClient) Update(character *Character) {
	// Send an update on character positions
	position := pb.Position{
		X:    float32(character.X),
		Y:    float32(character.Y),
		VelX: float32(character.forces[0]),
		VelY: float32(character.forces[1]),
	}

	update := &pb.ClientUpdate{
		UpdatePayload: &pb.ClientUpdate_P{
			P: &position,
		},
	}

	err := l.updateStream.Send(update)

	if err != nil {
		log.WithError(err).Error("Failed to send update")
		l.updateStream.CloseSend()
		l.updateStream = nil
	}

	// Load position updates waiting from any other characters
	go func() {
		for {
			serverUpdate, err := l.updateStream.Recv()
			if err != nil {
				log.WithError(err).Error("Unexpected error on update data stream")
				return
			}

			// TODO Apply position update
			// TODO Need to change position stream to deal with wrapped positions that include player data
			log.WithFields(log.Fields{
				"User":   serverUpdate.UserIdentifier,
				"Update": serverUpdate.String(),
			}).Debug("Update Received")
		}
	}()
}
