package handshake

import (
	"bytes"
	"errors"
	"strings"
	"testing"
)

func TestGenerateChallengeUses128BitsAndHexEncoding(t *testing.T) {
	input := make([]byte, ChallengeEntropyBytes)
	for i := range input { input[i] = byte(i) }
	got, err := GenerateChallengeFrom(bytes.NewReader(input))
	if err != nil { t.Fatal(err) }
	if len(got) != ChallengeEntropyBytes*2 { t.Fatalf("challenge length = %d", len(got)) }
	if got != "000102030405060708090a0b0c0d0e0f" { t.Fatalf("unexpected encoding: %q", got) }
	if strings.TrimSpace(got) != got { t.Fatal("challenge contains whitespace") }
	if err := validateExpectedChallenge(got); err != nil { t.Fatalf("generated challenge rejected: %v", err) }
}

func TestGenerateChallengeProducesDistinctValues(t *testing.T) {
	first, err := GenerateChallenge()
	if err != nil { t.Fatal(err) }
	second, err := GenerateChallenge()
	if err != nil { t.Fatal(err) }
	if first == second { t.Fatal("CSPRNG returned repeated challenge values") }
	if err := validateExpectedChallenge(first); err != nil { t.Fatal(err) }
	if err := validateExpectedChallenge(second); err != nil { t.Fatal(err) }
}

func TestGenerateChallengeFromRejectsMissingOrInsufficientEntropy(t *testing.T) {
	if _, err := GenerateChallengeFrom(nil); !errors.Is(err, ErrChallengeEntropy) { t.Fatalf("nil source: got %v", err) }
	if _, err := GenerateChallengeFrom(bytes.NewReader([]byte{1, 2, 3})); !errors.Is(err, ErrChallengeEntropy) { t.Fatalf("short source: got %v", err) }
}
