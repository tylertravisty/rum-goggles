package random

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
)

// Bytes creates an array of random bytes with specified size.
// If used to generate password salt, then size of 16 bytes (128 bits) is recommended.
func Bytes(size int) ([]byte, error) {
	if size < 0 {
		return nil, fmt.Errorf("random: size cannot be less than zero")
	}
	random := make([]byte, size)

	_, err := rand.Read(random)
	if err != nil {
		return nil, fmt.Errorf("random: error while reading random bytes: %v", err)
	}

	return random, nil
}

// String creates a Base64-encoded string of random bytes with specified length.
// If used to generate a session token, then length of 48 (36 bytes) is recommended.
func String(length int) (string, error) {
	b, err := Bytes(length)
	if err != nil {
		return "", fmt.Errorf("random: error while generating random bytes: %v", err)
	}

	str := base64.URLEncoding.EncodeToString(b)

	return str[:length], nil
}
