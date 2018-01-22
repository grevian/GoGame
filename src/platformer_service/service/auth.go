package service

import (
	"crypto/rsa"
	"fmt"
	"io/ioutil"

	log "github.com/Sirupsen/logrus"
	"github.com/dgrijalva/jwt-go"
	"golang.org/x/net/context"
	"google.golang.org/grpc/metadata"
)

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

func loadPublicKey(publicKeyPath string) (*rsa.PublicKey, error) {
	data, err := ioutil.ReadFile(publicKeyPath)
	if err != nil {
		return nil, fmt.Errorf("error reading the jwt public key: %v", err)
	}

	publickey, err := jwt.ParseRSAPublicKeyFromPEM(data)
	if err != nil {
		return nil, fmt.Errorf("error parsing the jwt public key: %s", err)
	}

	return publickey, nil
}
