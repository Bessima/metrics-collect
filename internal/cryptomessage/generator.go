package cryptomessage

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"
)

func GenerateKeys() {
	privateKey, _ := rsa.GenerateKey(rand.Reader, 2048)

	// Сохранение приватного ключа
	privFile, _ := os.Create("private.pem")
	defer privFile.Close()
	pem.Encode(privFile, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privateKey)})

	// Сохранение публичного ключа
	pubFile, _ := os.Create("public.pem")
	defer pubFile.Close()
	pem.Encode(pubFile, &pem.Block{Type: "RSA PUBLIC KEY", Bytes: x509.MarshalPKCS1PublicKey(&privateKey.PublicKey)})
}
