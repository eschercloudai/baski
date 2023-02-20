package sign

import (
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"log"
)

func EncodeKeys(pk *ecdsa.PrivateKey) ([]byte, []byte) {
	priv, err := x509.MarshalECPrivateKey(pk)
	if err != nil {
		log.Fatal(err)
	}
	privPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: priv,
	})

	pub, err := x509.MarshalPKIXPublicKey(&pk.PublicKey)
	if err != nil {
		log.Fatal(err)
	}
	pubPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pub,
	})

	return privPEM, pubPEM
}

func DecodePrivateKey(privPEM []byte) *ecdsa.PrivateKey {
	block, _ := pem.Decode(privPEM)
	x509Encoded := block.Bytes
	privateKey, err := x509.ParseECPrivateKey(x509Encoded)
	if err != nil {
		log.Fatal(err)
	}

	return privateKey
}

func DecodePublicKey(pubPEM []byte) *ecdsa.PublicKey {
	blockPub, _ := pem.Decode(pubPEM)
	x509EncodedPub := blockPub.Bytes
	genericPublicKey, err := x509.ParsePKIXPublicKey(x509EncodedPub)
	if err != nil {
		log.Fatal(err)
	}
	publicKey := genericPublicKey.(*ecdsa.PublicKey)

	return publicKey
}

func Sign(imgID string, privKey *ecdsa.PrivateKey) (string, error) {
	hash := sha256.Sum256([]byte(imgID))
	sign, err := ecdsa.SignASN1(rand.Reader, privKey, hash[:])
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(sign), nil
}

func Validate(imgID string, pubKey *ecdsa.PublicKey, sign string) (bool, error) {
	hash := sha256.Sum256([]byte(imgID))
	bSign, err := base64.StdEncoding.DecodeString(sign)
	if err != nil {
		return false, err
	}
	return ecdsa.VerifyASN1(pubKey, hash[:], bSign), nil
}
