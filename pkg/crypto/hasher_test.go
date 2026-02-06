package crypto

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestArgon2Hasher_Hash(t *testing.T) {
	t.Run("returns valid hash for valid password", func(t *testing.T) {
		hasher := NewArgon2Hasher()

		hash, err := hasher.Hash("Test@123!")

		require.NoError(t, err)
		assert.True(t, strings.HasPrefix(hash, "$argon2id$v=19$"))
		parts := strings.Split(hash, "$")
		assert.Len(t, parts, 6)
	})

	t.Run("generates unique hashes for same password", func(t *testing.T) {
		hasher := NewArgon2Hasher()

		hash1, err := hasher.Hash("Test@123!")
		require.NoError(t, err)

		hash2, err := hasher.Hash("Test@123!")
		require.NoError(t, err)

		assert.NotEqual(t, hash1, hash2)
	})

	t.Run("returns error for empty password", func(t *testing.T) {
		hasher := NewArgon2Hasher()

		hash, err := hasher.Hash("")

		assert.True(t, errors.Is(err, ErrEmptyPassword))
		assert.Empty(t, hash)
	})

	t.Run("uses custom params", func(t *testing.T) {
		params := &Argon2Params{
			Memory:      32 * 1024,
			Iterations:  1,
			Parallelism: 2,
			SaltLength:  16,
			KeyLength:   32,
		}
		hasher := NewArgon2HasherWithParams(params)

		hash, err := hasher.Hash("Test@123!")

		require.NoError(t, err)
		assert.Contains(t, hash, "m=32768,t=1,p=2")
	})
}

func TestArgon2Hasher_Verify(t *testing.T) {
	t.Run("returns true for matching password", func(t *testing.T) {
		hasher := NewArgon2Hasher()

		hash, err := hasher.Hash("Test@123!")
		require.NoError(t, err)

		match, err := hasher.Verify("Test@123!", hash)

		require.NoError(t, err)
		assert.True(t, match)
	})

	t.Run("returns false for wrong password", func(t *testing.T) {
		hasher := NewArgon2Hasher()

		hash, err := hasher.Hash("Test@123!")
		require.NoError(t, err)

		match, err := hasher.Verify("WrongPassword", hash)

		require.NoError(t, err)
		assert.False(t, match)
	})

	t.Run("returns error for malformed hash with wrong parts", func(t *testing.T) {
		hasher := NewArgon2Hasher()

		match, err := hasher.Verify("Test@123!", "not-a-valid-hash")

		assert.True(t, errors.Is(err, ErrInvalidHash))
		assert.False(t, match)
	})

	t.Run("returns error for unsupported algorithm", func(t *testing.T) {
		hasher := NewArgon2Hasher()

		match, err := hasher.Verify("Test@123!", "$bcrypt$v=19$m=65536,t=3,p=4$c2FsdA$a2V5")

		assert.True(t, errors.Is(err, ErrInvalidHash))
		assert.False(t, match)
	})

	t.Run("returns error for invalid version", func(t *testing.T) {
		hasher := NewArgon2Hasher()

		match, err := hasher.Verify("Test@123!", "$argon2id$v=99$m=65536,t=3,p=4$c2FsdA$a2V5")

		assert.True(t, errors.Is(err, ErrInvalidHash))
		assert.False(t, match)
	})

	t.Run("returns error for invalid base64 salt", func(t *testing.T) {
		hasher := NewArgon2Hasher()

		match, err := hasher.Verify("Test@123!", "$argon2id$v=19$m=65536,t=3,p=4$!!!invalid!!!$a2V5")

		assert.True(t, errors.Is(err, ErrInvalidHash))
		assert.False(t, match)
	})

	t.Run("returns error for invalid base64 key", func(t *testing.T) {
		hasher := NewArgon2Hasher()

		match, err := hasher.Verify("Test@123!", "$argon2id$v=19$m=65536,t=3,p=4$c2FsdA$!!!invalid!!!")

		assert.True(t, errors.Is(err, ErrInvalidHash))
		assert.False(t, match)
	})
}
