package util

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

var (
	ErrInvalidHash         = errors.New("the encoded hash is not in the correct format")
	ErrIncompatibleVersion = errors.New("incompatible version of argon2")
)

// Argon2Params defines the parameters for Argon2id hashing
type Argon2Params struct {
	Memory      uint32 // Memory in KiB
	Iterations  uint32 // Number of iterations
	Parallelism uint8  // Number of threads
	SaltLength  uint32 // Length of random salt in bytes
	KeyLength   uint32 // Length of generated hash in bytes
}

// DefaultArgon2Params returns recommended parameters for Argon2id
// These are OWASP recommended values for production use
func DefaultArgon2Params() *Argon2Params {
	return &Argon2Params{
		Memory:      64 * 1024, // 64 MB
		Iterations:  3,
		Parallelism: 2,
		SaltLength:  16,
		KeyLength:   32,
	}
}

// HashPassword generates an Argon2id hash of the password using default parameters
func HashPassword(password string) (string, error) {
	return HashPasswordWithParams(password, DefaultArgon2Params())
}

// HashPasswordWithParams generates an Argon2id hash of the password with custom parameters
func HashPasswordWithParams(password string, params *Argon2Params) (string, error) {
	// Generate a cryptographically secure random salt
	salt := make([]byte, params.SaltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}

	// Generate the hash
	hash := argon2.IDKey(
		[]byte(password),
		salt,
		params.Iterations,
		params.Memory,
		params.Parallelism,
		params.KeyLength,
	)

	// Encode the hash in a standard format
	// Format: $argon2id$v=19$m=65536,t=3,p=2$salt$hash
	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	encodedHash := fmt.Sprintf(
		"$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version,
		params.Memory,
		params.Iterations,
		params.Parallelism,
		b64Salt,
		b64Hash,
	)

	return encodedHash, nil
}

// VerifyPassword compares a password with an Argon2id hash
// Returns true if the password matches the hash, false otherwise
func VerifyPassword(password, encodedHash string) (bool, error) {
	// Extract the parameters, salt and derived key from the encoded hash
	params, salt, hash, err := decodeHash(encodedHash)
	if err != nil {
		return false, err
	}

	// Derive the key from the password using the same parameters
	otherHash := argon2.IDKey(
		[]byte(password),
		salt,
		params.Iterations,
		params.Memory,
		params.Parallelism,
		params.KeyLength,
	)

	// Use constant-time comparison to prevent timing attacks
	if subtle.ConstantTimeCompare(hash, otherHash) == 1 {
		return true, nil
	}

	return false, nil
}

// decodeHash extracts the parameters, salt, and hash from the encoded string
func decodeHash(encodedHash string) (*Argon2Params, []byte, []byte, error) {
	parts := strings.Split(encodedHash, "$")
	if len(parts) != 6 {
		return nil, nil, nil, ErrInvalidHash
	}

	// Check algorithm
	if parts[1] != "argon2id" {
		return nil, nil, nil, ErrInvalidHash
	}

	// Check version
	var version int
	if _, err := fmt.Sscanf(parts[2], "v=%d", &version); err != nil {
		return nil, nil, nil, err
	}
	if version != argon2.Version {
		return nil, nil, nil, ErrIncompatibleVersion
	}

	// Parse parameters
	params := &Argon2Params{}
	if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &params.Memory, &params.Iterations, &params.Parallelism); err != nil {
		return nil, nil, nil, err
	}

	// Decode salt
	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return nil, nil, nil, err
	}
	params.SaltLength = uint32(len(salt))

	// Decode hash
	hash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return nil, nil, nil, err
	}
	params.KeyLength = uint32(len(hash))

	return params, salt, hash, nil
}
