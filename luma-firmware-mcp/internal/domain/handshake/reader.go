package handshake

import (
	"context"
	"errors"
	"fmt"
	"time"
)

type LineSource interface {
	ReadLine(context.Context) ([]byte, error)
}

var (
	ErrInvalidTimeout = errors.New("handshake timeout must be positive")
	ErrNoLineSource = errors.New("handshake line source is nil")
)

// ReadIdentity consumes records from an injected source until it receives a valid
// identity response or fails. A valid ready record is skipped. This package has
// no serial adapter and does not authenticate the source.
func ReadIdentity(ctx context.Context, source LineSource, expectedChallenge string, timeout time.Duration) (Identity, error) {
	var zero Identity
	if source == nil {
		return zero, ErrNoLineSource
	}
	if timeout <= 0 {
		return zero, ErrInvalidTimeout
	}
	if err := validateExpectedChallenge(expectedChallenge); err != nil {
		return zero, err
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	result := make(chan readResult, 1)
	go func() {
		line, err := source.ReadLine(ctx)
		for err == nil {
			if IsReadyRecord(line) == nil {
				line, err = source.ReadLine(ctx)
				continue
			}
			identity, parseErr := ParseIdentity(line, expectedChallenge)
			result <- readResult{identity: identity, err: parseErr}
			return
		}
		result <- readResult{err: err}
	}()

	select {
	case <-ctx.Done():
		return zero, ctx.Err()
	case got := <-result:
		if got.err != nil {
			return zero, fmt.Errorf("read runtime identity: %w", got.err)
		}
		return got.identity, nil
	}
}

type readResult struct {
	identity Identity
	err error
}
