package main

import (
	"crypto/rsa"
	"crypto/tls"
	"io/ioutil"

	log "github.com/Sirupsen/logrus"
	"github.com/dgrijalva/jwt-go"
	"google.golang.org/grpc/credentials"

)

func loadTransportCredentials(tlsCert *string, tlsKey *string) credentials.TransportCredentials {
	// Load the credentials we'll use for transport security
	cert, err := tls.LoadX509KeyPair(*tlsCert, *tlsKey)
	if err != nil {
		log.WithFields(
			log.Fields{
				"tlsCert": *tlsCert,
				"tlsKey":  *tlsKey,
			}).WithError(err).Fatal("Could not load TLS Certificate")
	}

	transportCredentials := credentials.NewTLS(&tls.Config{
		InsecureSkipVerify: true,
		Certificates:       []tls.Certificate{cert},
	})

	return transportCredentials
}

func loadJWTPrivateKey(jwtPrivateKeyPath *string) *rsa.PrivateKey {
	// Load the private key that will be used to sign issued tokens
	privateKeyData, err := ioutil.ReadFile(*jwtPrivateKeyPath)
	if err != nil {
		log.WithError(err).Fatal("error reading the jwt private key file")
	}

	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(privateKeyData)
	if err != nil {
		log.WithError(err).Fatal("error parsing the jwt private key file")
	}

	return privateKey
}