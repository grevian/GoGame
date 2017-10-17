package client

import (
	"io/ioutil"

	"crypto/rsa"
	"fmt"
	"sync"

	"io"

	log "github.com/Sirupsen/logrus"
	"github.com/dgrijalva/jwt-go"
	pb_auth "github.com/grevian/GoGame/common/auth"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

const AUTHSERVER_ADDRESS = "localhost:8078"

type AuthServiceTokenFetcher struct {
	authClient     pb_auth.AuthServerClient
	jwtPublicKey   *rsa.PublicKey
	tlsCredentials credentials.TransportCredentials

	validToken *jwt.Token
	tokenMutex sync.RWMutex
}

func (n *AuthServiceTokenFetcher) updateToken(JWT *pb_auth.JWT) error {
	log.Debug("Updating Token")
	n.tokenMutex.Lock()
	defer n.tokenMutex.Unlock()

	validatedToken, err := jwt.Parse(JWT.Token, func(t *jwt.Token) (interface{}, error) {
		// Ensure the signing method is what we expect
		if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
			log.Errorf("Unexpected signing method: %v", t.Header["alg"])
			return nil, fmt.Errorf("invalid token")
		} else {
			return n.jwtPublicKey, nil
		}
	})

	if err != nil {
		log.WithError(err).Error("Error applying updated token")
		return err
	}

	n.validToken = validatedToken
	return nil
}

func NewAuthServiceTokenFetcher(userCredentials *pb_auth.Credentials, transportCredentials credentials.TransportCredentials, jwtPublicKeyPath *string) (credentials.PerRPCCredentials, error) {
	// Dial our authorization service
	conn, err := grpc.Dial(AUTHSERVER_ADDRESS,
		grpc.WithTransportCredentials(transportCredentials),
	)

	if err != nil {
		log.WithField("Authserver", AUTHSERVER_ADDRESS).WithError(err).Error("Could not access auth server")
		return nil, err
	}

	authServiceClient := pb_auth.NewAuthServerClient(conn)

	// Load our the public key used to validate our JWTs
	rawJWTPublicKey, err := ioutil.ReadFile(*jwtPublicKeyPath)
	if err != nil {
		log.WithField("jwtPublicKeyPath", *jwtPublicKeyPath).WithError(err).Error("Could not read public key")
		return nil, err
	}
	jwtPublicKey, err := jwt.ParseRSAPublicKeyFromPEM(rawJWTPublicKey)
	if err != nil {
		log.WithField("jwtPublicKeyPath", *jwtPublicKeyPath).WithError(err).Error("Could not parse public key")
		return nil, err
	}

	tokenStream, err := authServiceClient.Authorize(context.Background(), userCredentials)
	if err != nil {
		log.WithField("Username", userCredentials.Username).WithError(err).Error("Could not authorize user")
		return nil, err
	}

	firstToken, err := tokenStream.Recv()
	if err != nil {
		log.WithField("Username", userCredentials.Username).WithError(err).Error("Could not get first token")
		return nil, err
	}

	tokenFetcher := &AuthServiceTokenFetcher{
		authClient:     authServiceClient,
		jwtPublicKey:   jwtPublicKey,
		tlsCredentials: transportCredentials,
	}

	if err := tokenFetcher.updateToken(firstToken); err != nil {
		log.WithField("Username", userCredentials.Username).WithError(err).Error("Could not validate first token")
		return nil, err
	}

	// Continuously refresh the token whenever a new one is pushed from the server
	go func(a *AuthServiceTokenFetcher) {
		for {
			token, err := tokenStream.Recv()
			if err != nil {
				if err == io.EOF {
					log.WithField("Username", userCredentials.Username).WithError(err).Info("AuthService Token Fetcher stream was closed by the server")
				} else {
					log.WithField("Username", userCredentials.Username).WithError(err).Error("AuthService Token Fetcher stream unexpectedly failed")
				}
			}
			if err := a.updateToken(token); err != nil {
				log.WithField("Username", userCredentials.Username).WithError(err).Error("AuthService Token Fetcher failed to apply a token update")
			}
		}
	}(tokenFetcher)

	return tokenFetcher, nil
}

func (a AuthServiceTokenFetcher) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	a.tokenMutex.RLock()
	defer a.tokenMutex.RUnlock()
	return map[string]string{
		"authorization": a.validToken.Raw,
	}, nil
}

func (a AuthServiceTokenFetcher) RequireTransportSecurity() bool {
	return true
}
