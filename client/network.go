package client

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"image/color"
	"io/ioutil"

	log "github.com/Sirupsen/logrus"
	pb_auth "github.com/grevian/GoGame/common/auth"
	pb "github.com/grevian/GoGame/common/platformer"
	"github.com/hajimehoshi/ebiten"
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

	// The current users user identifier
	userIdentifier int32

	// A list of other players
	players map[int32]*Player

	networkClock int64
}

type Player struct {
	name     string
	position *Position

	op    *ebiten.DrawImageOptions
	image *ebiten.Image
}

func NewPlayer() *Player {
	image, _ := ebiten.NewImage(20, 20, ebiten.FilterLinear)
	image.Fill(color.RGBA{0, 0, 255, 128})

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Reset()
	op.GeoM.Translate(0, 0)

	return &Player{
		name:     "",
		position: &Position{X: 0, Y: 0},
		op:       op,
		image:    image,
	}
}

func (p *Player) Draw(screen *ebiten.Image) {
	p.op.GeoM.Reset()
	p.op.GeoM.Translate(p.position.X, p.position.Y)
	screen.DrawImage(p.image, p.op)
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

	n := &NetworkClient{
		tlsCredentials: transportCredentials,
		rpcCredentials: authTokenFetcher,
		gameClient:     gameClient,

		updateStream:   updateStream,
		userIdentifier: -1,
		players:        make(map[int32]*Player),

		networkClock: 0,
	}

	n.Start()

	return n, nil
}

func (l *NetworkClient) Start() {
	// Load position updates waiting from any other characters
	go func() {
		for {
			serverUpdate, err := l.updateStream.Recv()
			if err != nil {
				log.WithError(err).Error("Unexpected error on update data stream")
				return
			}

			switch serverUpdate.UpdatePayload.(type) {
			case *pb.ServerUpdate_C:
				l.processCommand(serverUpdate.UserIdentifier, serverUpdate.GetC())
			case *pb.ServerUpdate_P:
				l.processPositionUpdate(serverUpdate.UserIdentifier, serverUpdate.GetP())
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

func (l *NetworkClient) processCommand(userIdentifier int32, command *pb.Command) {
	switch command.Command {
	case pb.Command_QUIT:
		if userIdentifier == l.userIdentifier {
			// The server told you that you quit. Hope you were expecting that?

		} else {
			l.removePlayer(userIdentifier)
		}
		break
	case pb.Command_JOINED:
		if l.userIdentifier == -1 {
			// The first join we see should always be the server assigning us our ID
			log.WithField("user_id", userIdentifier).Info("We successfully joined")
			l.userIdentifier = userIdentifier
		} else {
			log.WithField("user_id", userIdentifier).Info("Another player joined")
			l.addPlayer(userIdentifier)
		}
	}
}

func (l *NetworkClient) processPositionUpdate(userIdentifier int32, position *pb.Position) {
	if l.userIdentifier == userIdentifier {
		log.WithField("update", position).Info("Our position was corrected by the server!")
		// TODO implement this
	} else {
		p, ok := l.players[userIdentifier]
		if !ok {
			log.WithField("user_id", userIdentifier).Error("Received an update for a player we haven't seen yet")
		}
		p.position.X = float64(position.X)
		p.position.Y = float64(position.Y)
	}
}

func (l *NetworkClient) addPlayer(userIdentifier int32) {
	// TODO Do a lookup of player information to display
	l.players[userIdentifier] = NewPlayer()
}

func (l *NetworkClient) removePlayer(userIdentifier int32) {
	// TODO Stop player processes before removing them
	delete(l.players, userIdentifier)
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
	}
}

func (l *NetworkClient) GetPlayers() map[int32]*Player {
	return l.players
}
