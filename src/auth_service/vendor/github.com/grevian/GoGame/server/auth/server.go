package auth

import (
	"crypto/rsa"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"time"

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

	TOKEN_EXPIRATION_DURATION time.Duration
	TOKEN_REFRESH_DURATION    time.Duration
	SKEW_NBF_DURATION         time.Duration
}

const (
	TOKEN_EXPIRATION_DURATION_STR = "15m"
	TOKEN_REFRESH_DURATION_STR    = "14m"
	SKEW_NBF_DURATION_STR         = "1m"
)

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

	// If any errors happen here, we're too screwed to even try and stop
	refreshDuration, _ := time.ParseDuration(TOKEN_REFRESH_DURATION_STR)
	expirationDuration, _ := time.ParseDuration(TOKEN_EXPIRATION_DURATION_STR)
	ndfDuration, _ := time.ParseDuration(SKEW_NBF_DURATION_STR)

	authServer := &AuthorizationServer{
		jwtPrivateKey:             privateKey,
		transportCredentials:      transportCredentials,
		TOKEN_REFRESH_DURATION:    refreshDuration,
		TOKEN_EXPIRATION_DURATION: expirationDuration,
		SKEW_NBF_DURATION:         ndfDuration,
	}

	return authServer, nil
}

func (a *AuthorizationServer) Serve(listener net.Listener) error {
	s := grpc.NewServer(grpc.Creds(a.transportCredentials))
	pb_auth.RegisterAuthServerServer(s, a)
	return s.Serve(listener)
}

func (a *AuthorizationServer) createTokenString(claims jwt.MapClaims) (string, error) {
	claims["exp"] = time.Now().Add(a.TOKEN_EXPIRATION_DURATION).Unix()
	claims["nbf"] = time.Now().Add(-a.SKEW_NBF_DURATION).Unix()
	claims["iat"] = time.Now().Unix()

	token := jwt.NewWithClaims(jwt.GetSigningMethod("RS512"), claims)
	tokenString, err := token.SignedString(a.jwtPrivateKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func (a *AuthorizationServer) authorizeUser(username string, password string) (map[string]string, error) {
	// TODO Look up user credentials in a database and confirm they are valid
	var userData = map[string]string{
		"user": "grevian",
	}

	return userData, nil
}

func (a *AuthorizationServer) Authorize(c *pb_auth.Credentials, tokenStream pb_auth.AuthServer_AuthorizeServer) error {
	userData, err := a.authorizeUser(c.Username, c.Password)
	if err != nil {
		log.WithError(err).WithField("Username", c.Username).Warn("User Authentication Failed")
		return errors.New("authorization failed")
	}

	// Add user data to the claims
	// TODO Ensure no sensitive data ends up here
	startingClaims := jwt.MapClaims{}
	for k, v := range userData {
		startingClaims[k] = v
	}

	// Send a refreshed token every time the users token is about to expire
	// TODO Also associate a semaphore or channel with the users session to support Logout cancelling this loop
	for {
		tokenString, err := a.createTokenString(startingClaims)
		if err != nil {
			log.WithError(err).Error("Failed to generate a signed token!")
			return err
		}

		log.WithFields(log.Fields{"user": "grevian", "tokenStr": tokenString}).Info("Issuing a token")
		err = tokenStream.Send(&pb_auth.JWT{Token: tokenString})
		if err != nil {
			log.WithError(err).Error("Failed to write a signed token to the tokenStream")
			return err
		}
		<-time.After(a.TOKEN_REFRESH_DURATION)
	}
	return nil
}

func (a *AuthorizationServer) Logout(context.Context, *pb_auth.JWT) (response *pb_auth.LogoutResponse, e error) {
	log.Error("Logout called, but is not implemented")
	return &pb_auth.LogoutResponse{}, nil
}
