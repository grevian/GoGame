package auth

import (
	"crypto/rsa"
	"fmt"
	"io/ioutil"

	log "github.com/Sirupsen/logrus"
	"github.com/dgrijalva/jwt-go"
	pb_auth "github.com/grevian/GoGame/common/auth"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type AuthorizationServer struct {
	jwtPrivateKey        *rsa.PrivateKey
	transportCredentials credentials.TransportCredentials
}

func NewAuthServer(rsaPrivateKeyPath *string, transportCredentials credentials.TransportCredentials) (*AuthorizationServer, error) {
	// Load the jwt private key that will be used to sign issued tokens
	privateKeyData, err := ioutil.ReadFile(*rsaPrivateKeyPath)
	if err != nil {
		return nil, fmt.Errorf("error reading the jwt public key: %v", err)
	}

	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(privateKeyData)
	if err != nil {
		return nil, fmt.Errorf("error parsing the jwt public key: %s", err)
	}

	s := grpc.NewServer(grpc.Creds(transportCredentials))
	authServer := &AuthorizationServer{
		jwtPrivateKey:        privateKey,
		transportCredentials: transportCredentials,
	}

	pb_auth.RegisterAuthServerServer(s, authServer)

	return authServer, nil
}

func (a *AuthorizationServer) Authorize(c *pb_auth.Credentials, tokenStream pb_auth.AuthServer_AuthorizeServer) error {
	log.Error("Authorize called, but is not implemented")
	// TODO Look up user credentials in a database, associate the tokenStream with a user session
	// TODO then issue a token to the stream for the user, and schedule another token whenever one is due to expire

	// TODO Add non-sensitive user/session information to the claims
	token := jwt.NewWithClaims(jwt.GetSigningMethod("RS512"), jwt.MapClaims{"": ""})
	tokenString, err := token.SignedString(a.jwtPrivateKey)
	if err != nil {
		log.WithError(err).Error("Failed to generate a signed token!")
	}

	jwtMessage := pb_auth.JWT{
		Token: tokenString,
	}

	err = tokenStream.Send(&jwtMessage)
	if err != nil {
		log.WithError(err).Error("Failed to write a signed token to the tokenStream")
	}

	return nil
}

func (a *AuthorizationServer) Logout(context.Context, *pb_auth.JWT) (response *pb_auth.LogoutResponse, e error) {
	log.Error("Logout called, but is not implemented")
	return &pb_auth.LogoutResponse{}, nil
}
