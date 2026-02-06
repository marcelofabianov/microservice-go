package crypto

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/marcelofabianov/fault"
	"golang.org/x/crypto/argon2"
)

var (
	ErrHashFailed    = fault.New("failed to hash password", fault.WithCode(fault.Internal))
	ErrInvalidHash   = fault.New("invalid password hash format", fault.WithCode(fault.Invalid))
	ErrEmptyPassword = fault.New("password cannot be empty", fault.WithCode(fault.Invalid))
)

type PasswordHasher interface {
	Hash(password string) (string, error)
	Verify(password, hash string) (bool, error)
}

type Argon2Params struct {
	Memory      uint32
	Iterations  uint32
	Parallelism uint8
	SaltLength  uint32
	KeyLength   uint32
}

func DefaultArgon2Params() *Argon2Params {
	return &Argon2Params{
		Memory:      64 * 1024,
		Iterations:  3,
		Parallelism: 4,
		SaltLength:  16,
		KeyLength:   32,
	}
}

type Argon2Hasher struct {
	params *Argon2Params
}

func NewArgon2Hasher() *Argon2Hasher {
	return &Argon2Hasher{params: DefaultArgon2Params()}
}

func NewArgon2HasherWithParams(params *Argon2Params) *Argon2Hasher {
	return &Argon2Hasher{params: params}
}

func (h *Argon2Hasher) Hash(password string) (string, error) {
	if password == "" {
		return "", ErrEmptyPassword
	}

	salt := make([]byte, h.params.SaltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", fault.Wrap(ErrHashFailed, "failed to generate salt",
			fault.WithCode(fault.Internal),
		)
	}

	key := argon2.IDKey(
		[]byte(password),
		salt,
		h.params.Iterations,
		h.params.Memory,
		h.params.Parallelism,
		h.params.KeyLength,
	)

	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Key := base64.RawStdEncoding.EncodeToString(key)

	hash := fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version,
		h.params.Memory,
		h.params.Iterations,
		h.params.Parallelism,
		b64Salt,
		b64Key,
	)

	return hash, nil
}

func (h *Argon2Hasher) Verify(password, hash string) (bool, error) {
	params, salt, key, err := decodeHash(hash)
	if err != nil {
		return false, err
	}

	otherKey := argon2.IDKey(
		[]byte(password),
		salt,
		params.Iterations,
		params.Memory,
		params.Parallelism,
		params.KeyLength,
	)

	return subtle.ConstantTimeCompare(key, otherKey) == 1, nil
}

func decodeHash(hash string) (*Argon2Params, []byte, []byte, error) {
	parts := strings.Split(hash, "$")
	if len(parts) != 6 {
		return nil, nil, nil, fault.Wrap(ErrInvalidHash, "expected 6 parts",
			fault.WithContext("parts", len(parts)),
		)
	}

	if parts[1] != "argon2id" {
		return nil, nil, nil, fault.Wrap(ErrInvalidHash, "unsupported algorithm",
			fault.WithContext("algorithm", parts[1]),
		)
	}

	var version int
	if _, err := fmt.Sscanf(parts[2], "v=%d", &version); err != nil {
		return nil, nil, nil, fault.Wrap(ErrInvalidHash, "failed to parse version")
	}
	if version != argon2.Version {
		return nil, nil, nil, fault.Wrap(ErrInvalidHash, "incompatible version",
			fault.WithContext("version", version),
		)
	}

	var memory uint32
	var iterations uint32
	var parallelism uint8
	if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &memory, &iterations, &parallelism); err != nil {
		return nil, nil, nil, fault.Wrap(ErrInvalidHash, "failed to parse parameters")
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return nil, nil, nil, fault.Wrap(ErrInvalidHash, "failed to decode salt")
	}

	key, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return nil, nil, nil, fault.Wrap(ErrInvalidHash, "failed to decode key")
	}

	return &Argon2Params{
		Memory:      memory,
		Iterations:  iterations,
		Parallelism: parallelism,
		SaltLength:  uint32(len(salt)), // #nosec G115 - salt length is always small (16-32 bytes)
		KeyLength:   uint32(len(key)),  // #nosec G115 - key length is always small (32-64 bytes)
	}, salt, key, nil
}

// Ensure Argon2Hasher implements PasswordHasher.
var _ PasswordHasher = (*Argon2Hasher)(nil)
