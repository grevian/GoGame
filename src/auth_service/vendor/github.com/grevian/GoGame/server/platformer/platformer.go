package platformer

import (
	"crypto/rsa"
	"fmt"
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

	level       *Level
	players     map[string]*Character
	identifiers int32
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
		identifiers:          1,
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

	// For now since we have no loaded data, just assign an auto-incrementing ID
	var user_id int32 = g.identifiers
	g.identifiers += 1

	// TODO Load more information about the user from the game service
	username := claims["user"].(string)
	username = fmt.Sprintf("%s-%d", username, user_id)

	log.WithField("user", username).Info("User Connected")
	character := NewNetworkCharacter(140, 310, &User{name: username, id: user_id}, updateStream)
	g.players[username] = character

	// Assign the new players id and let them know they've joined
	character.ServerUpdate(&pb.ServerUpdate{
		UserIdentifier: user_id,
		UpdatePayload: &pb.ServerUpdate_C{
			C: &pb.Command{
				Command: pb.Command_JOINED,
			},
		},
	})

	for _, player := range g.players {
		// Tell existing players about the new connection
		if player.user.id != user_id {
			player.ServerUpdate(&pb.ServerUpdate{
				UserIdentifier: user_id,
				UpdatePayload: &pb.ServerUpdate_C{
					C: &pb.Command{
						Command: pb.Command_JOINED,
					},
				},
			})

			player.ServerUpdate(&pb.ServerUpdate{
				UserIdentifier: user_id,
				UpdatePayload: &pb.ServerUpdate_P{
					P: &pb.Position{
						X: float32(character.position.X),
						Y: float32(character.position.Y),
					},
				},
			})

			// Tell the player about other existing players
			character.ServerUpdate(&pb.ServerUpdate{
				UserIdentifier: player.user.id,
				UpdatePayload: &pb.ServerUpdate_C{
					C: &pb.Command{
						Command: pb.Command_JOINED,
					},
				},
			})
			character.ServerUpdate(&pb.ServerUpdate{
				UserIdentifier: player.user.id,
				UpdatePayload: &pb.ServerUpdate_P{
					P: &pb.Position{
						X: float32(player.position.X),
						Y: float32(player.position.Y),
					},
				},
			})
		}
	}

	var updates []*pb.ServerUpdate
	for {
		select {
		// Server tick should try to match around 60fps clients
		case <-time.After(time.Millisecond * 16):
			updates = character.Tick(g.level, 500)

			// For all updates we're about to send, identify them as coming from this user
			// then distribute them to other users
			for i := range updates {
				updates[i].UserIdentifier = user_id

				for _, player := range g.players {
					if player.user.id != user_id {
						player.ServerUpdate(updates[i])
					}
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
