package sockets

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"fmt"
)

// GenerateKeys generates an 2048 bit RSA keypair using a cryptographically
// secure random number generator selected by crypto/rand.Reader.
//
// Returns a pointer to the private key and a copy of the public key.
func GenerateKeys() (*rsa.PrivateKey, rsa.PublicKey) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		panic(err)
	}

	return privateKey, privateKey.PublicKey
}

// Encrypt will use the given public key to encrypt the given plaintext using
// RSA-OEAP (Optimal Asymmetric Encryption Padding) encryption. The plaintext is
// hashed with SHA256 and salted with crypto/rand.Reader generated bits.
//
// Returns a ciphertext byte array.
func Encrypt(publicKey rsa.PublicKey, plainText string) []byte {
	encryptedBytes, err := rsa.EncryptOAEP(
		sha256.New(),
		rand.Reader,
		&publicKey,
		[]byte(plainText),
		nil)
	if err != nil {
		panic(err)
	}

	return encryptedBytes
}

// Decrypt uses the given private key to decrypt the given bytes. It is assumed
// the bytes were encrypted with RSA-OEAP and hashed with SHA256.
//
// Returns a decrypted string of plaintext.
func Decrypt(privateKey *rsa.PrivateKey, encryptedBytes []byte) string {
	decryptedBytes, err := privateKey.Decrypt(
		nil,
		encryptedBytes,
		&rsa.OAEPOptions{Hash: crypto.SHA256})
	if err != nil {
		panic(err)
	}

	return string(decryptedBytes)
}

// Sign uses the given private key to sign a checksummed hash generated from the
// given plaintext. The hash is generated with SHA256 and salted with
// crypto/rand.Reader, and the signature is generated with RSASSA-PSS.
//
// Returns a signature byte array.
func Sign(privateKey *rsa.PrivateKey, plainText string) []byte {
	msgHashSum := generateHashSum(plainText)

	signature, err := rsa.SignPSS(
		rand.Reader,
		privateKey,
		crypto.SHA256,
		msgHashSum,
		nil)
	if err != nil {
		panic(err)
	}

	return signature
}

// Verify verifies the given plaintext by first generating its checksummed hash,
// then verifying that hash with the given public key. The hash is generated
// with SHA256 and salted with crypto/rand.Reader, and the signature is
// generated with RSASSA-PSS.
//
// Panics if the verification fails.
func Verify(publicKey rsa.PublicKey, plainText string, signature []byte) {
	msgHashSum := generateHashSum(plainText)

	err := rsa.VerifyPSS(
		&publicKey,
		crypto.SHA256,
		msgHashSum,
		signature,
		nil)
	if err != nil {
		panic(err)
	}
}

// TestRSA runs a demonstration of RSA encryption and signing.
func TestRSA() {
	// server creates keys
	privateKey, publicKey := GenerateKeys()
	fmt.Println("Server generates keys")

	// server sends public key to client

	// client creates plaintext
	plainText := "hello there"
	fmt.Println("Client creates plaintext: " + plainText)

	// client encrypts plaintext
	encryptedBytes := Encrypt(publicKey, plainText)
	fmt.Println("Client generates ciphertext")

	signature := Sign(privateKey, plainText)
	fmt.Println("Client generates signature")

	// client sends ciphertext

	// server decrypts ciphertext
	decryptedPlainText := Decrypt(privateKey, encryptedBytes)
	fmt.Println("Server decrypts plaintext: " + decryptedPlainText)

	// server verifies message
	Verify(publicKey, decryptedPlainText, signature)
	fmt.Println("Server verifies plaintext and signature successfully")

	// server verifies a message modified by a man-in-the-middle
	modifiedMessage := decryptedPlainText +
		", please send me your credit card details"
	fmt.Println("Server verifies a man-in-the-middle's modified message: " +
		modifiedMessage)
	Verify(publicKey, modifiedMessage, signature)
}

// generateHashSum generates a checksummed hash of the given plaintext using
// SHA256.
//
// Returns a checksummed hash in a byte array.
func generateHashSum(plainText string) []byte {
	msgHash := sha256.New()
	_, err := msgHash.Write([]byte(plainText))
	if err != nil {
		panic(err)
	}
	return msgHash.Sum(nil)
}
