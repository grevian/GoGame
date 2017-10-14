package platformer

import (
	"crypto/rsa"
	"net"

	log "github.com/Sirupsen/logrus"
	pb "github.com/grevian/GoGame/common/platformer"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type GameServer struct {
	jwtPublicKey         *rsa.PublicKey
	transportCredentials credentials.TransportCredentials
}

func NewGameServer(publicKeyPath string, transportCredentials credentials.TransportCredentials) (*GameServer, error) {
	publicKey, err := loadPublicKey(publicKeyPath)
	if err != nil {
		log.WithError(err).Fatal("Could not load public key")
	}

	return &GameServer{
		jwtPublicKey:         publicKey,
		transportCredentials: transportCredentials,
	}, nil
}

func (g *GameServer) Serve(listener net.Listener) error {
	// Create a grpc server that will validate transport credentials
	s := grpc.NewServer(grpc.Creds(g.transportCredentials))

	// Register our game service with the grpc server
	pb.RegisterGameServerServer(s, g)

	return s.Serve(listener)
}

func (g *GameServer) PositionUpdates(server pb.GameServer_PositionUpdatesServer) error {
	log.Error("PositionUpdates called, but is not yet implemented")
	return nil
}

func (g *GameServer) CommandUpdates(server pb.GameServer_CommandUpdatesServer) error {
	log.Error("CommandUpdates called, but is not yet implemented")
	return nil
}

func (g *GameServer) UserInformation(ctx context.Context, data *pb.UserData) (*pb.UserData, error) {
	log.Error("UserInformation called, but is not yet implemented")
	return nil, nil
}