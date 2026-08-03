package users

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"

	"golang.org/x/crypto/argon2"
)

func HashPassword(plaintext string) (string, error) {
	const (
		time    = 2
		memory  = 19 * 1024 // 19 MiB as KiB
		threads = 1
		keyLen  = 32
	)

	// Generate a random salt
	var salt = make([]byte, 16)
	_, err := rand.Read(salt)
	if err != nil { // RNG failure must abort
		return "", err
	}

	var hash = argon2.IDKey([]byte(plaintext), salt, time, memory, threads, keyLen)

	// Base64-encode the raw bytes
	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	// Bundle everything into the standard PHC / modular-crypt format.
	encoded := fmt.Sprintf(
		"$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version, memory, time, threads, b64Salt, b64Hash,
	)

	return encoded, nil
}
