package platformer

import (
	"crypto/rsa"
	"net"

	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/dgrijalva/jwt-go"
	pb "github.com/grevian/GoGame/common/platformer"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type GameServer struct {
	jwtPublicKey         *rsa.PublicKey
	transportCredentials credentials.TransportCredentials

	level   *Level
	players map[string]*Character
}

func NewGameServer(publicKeyPath string, transportCredentials credentials.TransportCredentials) (*GameServer, error) {
	publicKey, err := loadPublicKey(publicKeyPath)
	if err != nil {
		log.WithError(err).Fatal("Could not load public key")
	}

	return &GameServer{
		players:              make(map[string]*Character),
		jwtPublicKey:         publicKey,
		transportCredentials: transportCredentials,
		level:                NewLevel(),
	}, nil
}

func (g *GameServer) Serve(listener net.Listener) error {
	// Create a grpc server that will validate transport credentials
	s := grpc.NewServer(grpc.Creds(g.transportCredentials))

	// Register our game service with the grpc server
	pb.RegisterGameServerServer(s, g)

	return s.Serve(listener)
}

func (g *GameServer) Connect(updateStream pb.GameServer_ConnectServer) error {
	token, err := g.validateTokenFromContext(updateStream.Context())
	if err != nil {
		log.WithError(err).Error("Invalid Token")
		return err
	}
	claims := token.Claims.(jwt.MapClaims)

	// TODO Load more information about the user from the game service
	username := claims["user"].(string)
	var user_id int32 = 0

	log.WithField("user", username).Info("User Connected")
	character := NewNetworkCharacter(140, 310, &User{name: username}, updateStream)
	g.players[username] = character

	var updates []*pb.ServerUpdate
	for {
		select {
		case <-time.After(time.Millisecond * 16):
			updates = character.Tick(g.level, 500)

			// For all updates we're about to send, identify them as coming from this user
			// then distribute them to other users
			for i := range updates {
				updates[i].UserIdentifier = user_id

				for _, player := range g.players {
					player.ServerUpdate(updates[i])
				}
			}

			_ = updates
		}
	}
}

func (g *GameServer) UserInformation(ctx context.Context, data *pb.UserData) (*pb.UserData, error) {
	log.Error("UserInformation called, but is not yet implemented")
	return nil, nil
}
