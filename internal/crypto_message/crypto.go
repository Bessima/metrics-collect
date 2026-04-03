package crypto_message

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"os"
)

func GetPublicKey(publicKeyPath string) (*rsa.PublicKey, error) {
	pubKeyFile, err := os.ReadFile(publicKeyPath)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(pubKeyFile)
	pubKey, err := x509.ParsePKCS1PublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	return pubKey, nil
}

func GetPrivateKey(privateKeyPath string) (*rsa.PrivateKey, error) {
	privKeyFile, err := os.ReadFile(privateKeyPath)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(privKeyFile)
	privKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	return privKey, nil
}

func EncryptMessage(message []byte, publicKey *rsa.PublicKey) ([]byte, error) {
	ciphertext, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, publicKey, message, nil)
	if err != nil {
		return nil, err
	}
	return ciphertext, nil
}

func DecryptMessage(ciphertext []byte, privateKey *rsa.PrivateKey) ([]byte, error) {
	plaintext, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, privateKey, ciphertext, nil)
	if err != nil {
		return nil, err
	}
	return plaintext, nil
}
