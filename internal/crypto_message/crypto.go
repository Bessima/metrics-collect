package crypto_message

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/binary"
	"encoding/pem"
	"fmt"
	"io"
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

// EncryptMessage шифрует сообщение произвольного размера через гибридное шифрование:
// генерирует случайный AES-256 ключ, шифрует данные AES-GCM,
// AES-ключ шифрует RSA-OAEP.
// Формат: [2 байта: длина RSA-блока][RSA(aes_key)][nonce][AES-GCM(данные)]
func EncryptMessage(message []byte, publicKey *rsa.PublicKey) ([]byte, error) {
	aesKey := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, aesKey); err != nil {
		return nil, fmt.Errorf("generate aes key: %w", err)
	}

	encryptedKey, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, publicKey, aesKey, nil)
	if err != nil {
		return nil, fmt.Errorf("encrypt aes key: %w", err)
	}

	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return nil, fmt.Errorf("create aes cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("create gcm: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("generate nonce: %w", err)
	}

	ciphertext := gcm.Seal(nonce, nonce, message, nil)

	result := make([]byte, 2+len(encryptedKey)+len(ciphertext))
	binary.BigEndian.PutUint16(result[:2], uint16(len(encryptedKey)))
	copy(result[2:], encryptedKey)
	copy(result[2+len(encryptedKey):], ciphertext)

	return result, nil
}

// DecryptMessage расшифровывает сообщение, зашифрованное через EncryptMessage.
func DecryptMessage(data []byte, privateKey *rsa.PrivateKey) ([]byte, error) {
	if len(data) < 2 {
		return nil, fmt.Errorf("data too short")
	}

	keyLen := int(binary.BigEndian.Uint16(data[:2]))
	if len(data) < 2+keyLen {
		return nil, fmt.Errorf("data too short for encrypted key")
	}

	encryptedKey := data[2 : 2+keyLen]
	ciphertext := data[2+keyLen:]

	aesKey, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, privateKey, encryptedKey, nil)
	if err != nil {
		return nil, fmt.Errorf("decrypt aes key: %w", err)
	}

	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return nil, fmt.Errorf("create aes cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("create gcm: %w", err)
	}

	if len(ciphertext) < gcm.NonceSize() {
		return nil, fmt.Errorf("ciphertext too short")
	}

	nonce := ciphertext[:gcm.NonceSize()]
	ciphertext = ciphertext[gcm.NonceSize():]

	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("decrypt data: %w", err)
	}

	return plaintext, nil
}
