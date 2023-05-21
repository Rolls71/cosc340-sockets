package sockets

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"
)

func GenerateAESKey() []byte {
	key := make([]byte, 32)
	_, err := io.ReadFull(rand.Reader, key)
	if err != nil {
		fmt.Println(err)
	}
	return key
}

func EncryptAES(key []byte, plaintext string) ([]byte, bool) {
	c, err := aes.NewCipher(key)
	if err != nil {
		fmt.Println(err)
		return []byte{}, false
	}

	// gcm or Galois/Counter Mode, is a mode of operation
	// for symmetric key cryptographic block ciphers
	// - https://en.wikipedia.org/wiki/Galois/Counter_Mode
	gcm, err := cipher.NewGCM(c)
	if err != nil {
		fmt.Println(err)
		return []byte{}, false
	}

	// creates a new byte array the size of the nonce
	// which must be passed to Seal
	nonce := make([]byte, gcm.NonceSize())
	// populates our nonce with a cryptographically secure
	// random sequence
	_, err = io.ReadFull(rand.Reader, nonce)
	if err != nil {
		fmt.Println(err)
	}

	// here we encrypt our text using the Seal function
	// Seal encrypts and authenticates plaintext, authenticates the
	// additional data and appends the result to dst, returning the updated
	// slice. The nonce must be NonceSize() bytes long and unique for all
	// time, for a given key.
	return gcm.Seal(nonce, nonce, []byte(plaintext), nil), true
}

func DecryptAES(key, encryptedBytes []byte) ([]byte, bool) {
	c, err := aes.NewCipher(key)
	if err != nil {
		fmt.Println(err)
		return []byte{}, false
	}

	gcm, err := cipher.NewGCM(c)
	if err != nil {
		fmt.Println(err)
		return []byte{}, false
	}

	nonceSize := gcm.NonceSize()
	if len(encryptedBytes) < nonceSize {
		fmt.Println(err)
		return []byte{}, false
	}

	nonce, encryptedBytes := encryptedBytes[:nonceSize], encryptedBytes[nonceSize:]
	decryptedBytes, err := gcm.Open(nil, nonce, encryptedBytes, nil)
	if err != nil {
		fmt.Println(err)
		return []byte{}, false
	}
	return decryptedBytes, true
}

func TestAES() {
	plaintext := "Hello there"
	fmt.Println("Plaintext: " + plaintext)
	key := GenerateAESKey()

	encryptedBytes, ok := EncryptAES(key, plaintext)
	if !ok {
		panic("Failed to encrypt!")
	}
	fmt.Println("Ciphertext:\n" + string(encryptedBytes))

	transitString := string(encryptedBytes)

	decryptedBytes, ok := DecryptAES(key, []byte(transitString))
	if !ok {
		panic("Failed to decrypt!")
	}
	fmt.Println("Decrypted Plaintext: " + string(decryptedBytes))
}
