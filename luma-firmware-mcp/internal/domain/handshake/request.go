package handshake

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

var (
	ErrNoLineWriter = errors.New("handshake line writer is nil")
	ErrRequestWrite = errors.New("write handshake identity request")
)

type LineWriter interface {
	WriteLine(context.Context, []byte) error
}

type identifyRequest struct {
	Protocol        string `json:"protocol"`
	ProtocolVersion int    `json:"protocol_version"`
	Command         string `json:"command"`
	Challenge       string `json:"challenge"`
}

// EncodeIdentifyRequest serializes one protocol-v1 identify command without a
// line terminator. The transport owns framing and must write exactly one line.
func EncodeIdentifyRequest(challenge string) ([]byte, error) {
	if err := validateExpectedChallenge(challenge); err != nil {
		return nil, err
	}
	return json.Marshal(identifyRequest{
		Protocol: ProtocolName, ProtocolVersion: ProtocolVersion,
		Command: "identify", Challenge: challenge,
	})
}

// RequestIdentity sends a fresh-for-this-attempt challenge supplied by the
// caller, then waits for the matching identity under one overall deadline.
// The source is closed on write failure or by ReadIdentity once reading starts.
// It does not open ports, reset hardware, or authenticate the device.
func RequestIdentity(ctx context.Context, source interface {
	LineSource
	LineWriter
}, challenge string, timeout time.Duration) (Identity, error) {
	var zero Identity
	if source == nil {
		return zero, ErrNoLineSource
	}
	if timeout <= 0 {
		return zero, ErrInvalidTimeout
	}
	if ctx == nil {
		return zero, ErrNoContext
	}
	if err := validateExpectedChallenge(challenge); err != nil {
		return zero, err
	}
	request, err := EncodeIdentifyRequest(challenge)
	if err != nil {
		return zero, err
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	framed := append(request, '\n')
	if err := source.WriteLine(ctx, framed); err != nil {
		_ = source.Close()
		return zero, fmt.Errorf("%w: %w", ErrRequestWrite, err)
	}
	return ReadIdentity(ctx, source, challenge, timeout)
}
