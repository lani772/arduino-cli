package handshake

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

var (
	ErrNoTransport        = errors.New("handshake transport is nil")
	ErrNoOperatorConsent  = errors.New("explicit operator consent is required to open a serial transport")
	ErrNoPort             = errors.New("an exact serial port must be selected")
	ErrTransportOpen      = errors.New("open handshake transport")
)

// Transport opens a source for one explicitly selected port. Implementations
// must honor context cancellation while opening and must not guess another port.
type Transport interface {
	Open(ctx context.Context, port string) (LineSource, error)
}

// OpenSource validates consent and the exact port before delegating to a
// transport. It does not discover ports, reset a board, or authorize hardware
// writes. Ownership of the returned source transfers to the caller, which must
// close it; ReadIdentity closes sources once reading begins.
func OpenSource(ctx context.Context, transport Transport, port string, operatorConsent bool) (LineSource, error) {
	if ctx == nil {
		return nil, ErrNoContext
	}
	if transport == nil {
		return nil, ErrNoTransport
	}
	if !operatorConsent {
		return nil, ErrNoOperatorConsent
	}
	if strings.TrimSpace(port) == "" || strings.TrimSpace(port) != port {
		return nil, ErrNoPort
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	source, err := transport.Open(ctx, port)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrTransportOpen, err)
	}
	if source == nil {
		return nil, fmt.Errorf("%w: transport returned a nil source", ErrTransportOpen)
	}
	return source, nil
}
