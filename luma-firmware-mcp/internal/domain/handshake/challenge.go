package handshake

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"io"
)

const ChallengeEntropyBytes = 16

var ErrChallengeEntropy = errors.New("generate handshake challenge entropy")

// GenerateChallenge creates a fresh challenge from the operating system CSPRNG.
// The 128-bit random value is encoded as lowercase hexadecimal for JSON transport.
func GenerateChallenge() (string, error) {
	return GenerateChallengeFrom(rand.Reader)
}

// GenerateChallengeFrom is exposed for deterministic failure testing. Production
// callers should use GenerateChallenge, not provide their own entropy source.
func GenerateChallengeFrom(source io.Reader) (string, error) {
	if source == nil {
		return "", ErrChallengeEntropy
	}
	var raw [ChallengeEntropyBytes]byte
	if _, err := io.ReadFull(source, raw[:]); err != nil {
		return "", errors.Join(ErrChallengeEntropy, err)
	}
	return hex.EncodeToString(raw[:]), nil
}
