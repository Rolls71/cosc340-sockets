package sockets

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"fmt"
	"math/big"
	"strconv"
	"strings"
)

// GenerateRSAKeys generates a 2048 bit RSA keypair using a cryptographically
// secure random number generator selected by crypto/rand.Reader.
//
// Returns a pointer to the private key and a copy of the public key.
func GenerateRSAKeys() (*rsa.PrivateKey, rsa.PublicKey) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		panic(err)
	}

	return privateKey, privateKey.PublicKey
}

// EncryptRSA will use the given public key to encrypt the given plaintext using
// RSA-OEAP (Optimal Asymmetric Encryption Padding) encryption. The plaintext is
// hashed with SHA256 and salted with crypto/rand.Reader generated bits.
//
// Returns a ciphertext byte array and true if successful. Otherwise returns an
// empty array and a false value.
func EncryptRSA(publicKey rsa.PublicKey, plainText string) ([]byte, bool) {
	encryptedBytes, err := rsa.EncryptOAEP(
		sha256.New(),
		rand.Reader,
		&publicKey,
		[]byte(plainText),
		nil)
	if err != nil {
		return []byte{}, false
	}

	return encryptedBytes, true
}

// DecryptRSA uses the given private key to decrypt the given bytes. It is assumed
// the bytes were encrypted with RSA-OEAP and hashed with SHA256.
//
// Returns a plaintext byte array and true if successful. Otherwise returns an
// empty byte array and false
func DecryptRSA(privateKey *rsa.PrivateKey, encryptedBytes []byte) ([]byte, bool) {
	decryptedBytes, err := privateKey.Decrypt(
		nil,
		encryptedBytes,
		&rsa.OAEPOptions{Hash: crypto.SHA256})
	if err != nil {
		return []byte{}, false
	}

	return decryptedBytes, true
}

// SignRSA uses the given private key to sign a checksummed hash generated from the
// given plaintext. The hash is generated with SHA256 and salted with
// crypto/rand.Reader, and the signature is generated with RSASSA-PSS.
//
// Returns a signature byte array.
func SignRSA(privateKey *rsa.PrivateKey, plainText string) []byte {
	msgHashSum := generateHashSumRSA(plainText)

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

// VerifyRSA verifies the given plaintext by first generating its checksummed hash,
// then verifying that hash with the given public key. The hash is generated
// with SHA256 and salted with crypto/rand.Reader, and the signature is
// generated with RSASSA-PSS.
//
// Returns true if successful, otherwise returns false.
func VerifyRSA(
	publicKey rsa.PublicKey,
	plainText string,
	signature []byte,
) bool {
	msgHashSum := generateHashSumRSA(plainText)

	err := rsa.VerifyPSS(
		&publicKey,
		crypto.SHA256,
		msgHashSum,
		signature,
		nil)
	if err != nil {
		fmt.Println(err)
		return false
	}
	return true
}

// RSAKeyToString converts a given RSA public key to a string. The string
// contains the modulus and the exponent separated by a '-' character.
// e.g. "[modulus]-[exponent]"
//
// Returns the created string
func RSAKeyToString(publicKey rsa.PublicKey) string {
	return publicKey.N.String() + "-" + strconv.Itoa(publicKey.E)
}

// StringToRSAKey converts a given string into an RSA public key. The string
// must contain the modulus followed by the exponent separated by a '-'
// character.
// e.g. "[modulus]-[exponent]"
//
// Returns the created crypto/rsa.PublicKey
func StringToRSAKey(publicKey string) (rsa.PublicKey, bool) {
	strs := strings.Split(publicKey, "-")
	bi := big.NewInt(0)
	_, ok := bi.SetString(strs[0], 10)
	if !ok {
		fmt.Println("ERROR: Failed to convert public key to big int")
		return rsa.PublicKey{}, false
	}
	exponent, err := strconv.Atoi(strs[1])
	if err != nil {
		fmt.Println("ERROR: Failed to convert exponent to int")
		return rsa.PublicKey{}, false
	}
	return rsa.PublicKey{N: bi, E: exponent}, true
}

// TestRSA runs a demonstration of RSA encryption and signing.
func TestRSA() {
	// server creates keys
	privateKey, publicKey := GenerateRSAKeys()
	fmt.Println("Server generates keys")

	// server sends public key to client
	str := RSAKeyToString(publicKey)
	fmt.Println(str)
	_, ok := StringToRSAKey(str)
	if ok {
		fmt.Println("Public key sent!")
	} else {
		fmt.Println("Conversions unsuccessful, keys failed to send")
	}

	// client creates plaintext
	plainText := "hello there"
	fmt.Println("Client creates plaintext: " + plainText)

	// client encrypts plaintext
	encryptedBytes, ok := EncryptRSA(publicKey, plainText)
	if ok {
		fmt.Println("Client generates ciphertext")
	} else {
		fmt.Println("Client failed to generate ciphertext")
	}

	signature := SignRSA(privateKey, plainText)
	fmt.Println("Client generates signature")

	// client sends ciphertext

	// server decrypts ciphertext
	decryptedPlainText, ok := DecryptRSA(privateKey, encryptedBytes)
	if ok {
		fmt.Println("Server decrypts plaintext: " + string(decryptedPlainText))
	} else {
		fmt.Println("Server failed to decrypt plaintext")
	}

	// server verifies message
	if VerifyRSA(publicKey, string(decryptedPlainText), signature) {
		fmt.Println("Server verifies plaintext and signature successfully")
	}

	// server verifies a message modified by a man-in-the-middle
	modifiedMessage := string(decryptedPlainText) +
		", please send me your credit card details"
	fmt.Println("Server verifies a man-in-the-middle's modified message: " +
		modifiedMessage)
	fmt.Println(VerifyRSA(publicKey, modifiedMessage, signature))
}

// generateHashSumRSA generates a checksummed hash of the given plaintext using
// SHA256.
//
// Returns a checksummed hash in a byte array.
func generateHashSumRSA(plainText string) []byte {
	msgHash := sha256.New()
	_, err := msgHash.Write([]byte(plainText))
	if err != nil {
		panic(err)
	}
	return msgHash.Sum(nil)
}
