package sockets

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"
)

// GenerateAESKey generates a 32-byte key that can be used to create an AES-256
// cipher. Bytes are randomly selected by crypto/rand.Reader until a 32 byte
// array is full.
//
// Returns a 32-long byte array
func GenerateAESKey() []byte {
	key := make([]byte, 32)
	_, err := io.ReadFull(rand.Reader, key)
	if err != nil {
		fmt.Println(err)
	}
	return key
}

// EncryptAES takes a given key and plaintext and produces an encrypted byte
// array. The key is first used to create an AES-256 cipher in
// Galois/Counter mode. A series of random bytes are chosen by
// crypto/rand.Reader and used to create an authentication tag. Finally, an
// encrypted array of bytes is produced with a true value to indicate success.
//
// Returns an array of bytes and true if successful. Otherwise, returns an
// empty byte array and false.
func EncryptAES(key []byte, plaintext string) ([]byte, bool) {
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

	randBytes := make([]byte, gcm.NonceSize())
	_, err = io.ReadFull(rand.Reader, randBytes)
	if err != nil {
		fmt.Println(err)
	}

	return gcm.Seal(randBytes, randBytes, []byte(plaintext), nil), true
}

// DecryptAES takes two byte arrays, a key and some encrypted data, and produces
// a decrypted byte array. It is assumed that the key is 32 bytes long and the
// data has been encrypted by an AES-256 cipher in Galois/Counter mode.
//
// Returns a decrypted byte array and true if successful, otherwise an empty
// byte array and false.
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

// TestRSA runs a demonstration of AES-256 encryption in Galois/Counter mode.
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
