package service

import (
	"crypto/rsa"
	"errors"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/dgrijalva/jwt-go"
	"golang.org/x/net/context"

	pbAuth "github.com/grevian/GoGame/common/auth"
)

type AuthorizationServer struct {
	jwtPrivateKey *rsa.PrivateKey
	stop          chan bool

	TOKEN_EXPIRATION_DURATION time.Duration
	TOKEN_REFRESH_DURATION    time.Duration
	SKEW_NBF_DURATION         time.Duration
}

const (
	TOKEN_EXPIRATION_DURATION_STR = "15m"
	TOKEN_REFRESH_DURATION_STR    = "14m"
	SKEW_NBF_DURATION_STR         = "1m"
)

func NewAuthServer(privateKey *rsa.PrivateKey) (*AuthorizationServer, error) {
	refreshDuration, _ := time.ParseDuration(TOKEN_REFRESH_DURATION_STR)
	expirationDuration, _ := time.ParseDuration(TOKEN_EXPIRATION_DURATION_STR)
	ndfDuration, _ := time.ParseDuration(SKEW_NBF_DURATION_STR)

	authServer := &AuthorizationServer{
		jwtPrivateKey: privateKey,
		stop:          make(chan bool),
		TOKEN_REFRESH_DURATION:    refreshDuration,
		TOKEN_EXPIRATION_DURATION: expirationDuration,
		SKEW_NBF_DURATION:         ndfDuration,
	}

	return authServer, nil
}

func (a *AuthorizationServer) GracefulStop() {
	// Tell requests in progress to wrap up
	close(a.stop)
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

func (a *AuthorizationServer) Authorize(c *pbAuth.Credentials, tokenStream pbAuth.AuthServer_AuthorizeServer) error {
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
	for {
		tokenString, err := a.createTokenString(startingClaims)
		if err != nil {
			log.WithError(err).Error("Failed to generate a signed token!")
			return err
		}

		log.WithFields(log.Fields{"user": "grevian", "tokenStr": tokenString}).Info("Issuing a token")
		err = tokenStream.Send(&pbAuth.JWT{Token: tokenString})
		if err != nil {
			log.WithError(err).Error("Failed to write a signed token to the tokenStream")
			return err
		}
		select {
		case <-time.After(a.TOKEN_REFRESH_DURATION):
			continue
		case _ = <-a.stop:
			return errors.New("authorization service was stopped")
		}

	}
	return nil
}

func (a *AuthorizationServer) Logout(context.Context, *pbAuth.JWT) (response *pbAuth.LogoutResponse, e error) {
	log.Error("Logout called, but is not implemented")
	return &pbAuth.LogoutResponse{}, nil
}
