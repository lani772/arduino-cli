package handshake

import (
	"context"
	"errors"
	"io"
	"testing"
	"time"
)

type fakeLineSource struct {
	lines [][]byte
	index int
	err error
}

func (f *fakeLineSource) ReadLine(ctx context.Context) ([]byte, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	if f.index < len(f.lines) {
		line := f.lines[f.index]
		f.index++
		return line, nil
	}
	if f.err != nil {
		return nil, f.err
	}
	return nil, io.EOF
}

type blockingLineSource struct{}

func (blockingLineSource) ReadLine(ctx context.Context) ([]byte, error) {
	<-ctx.Done()
	return nil, ctx.Err()
}

func TestReadIdentitySkipsReadyThenAcceptsIdentity(t *testing.T) {
	ready := []byte(`{"protocol":"luma.runtime","protocol_version":1,"event":"ready","firmware_version":"1.0.0","project_id":"luma-test","device_id":"esp32-test","uptime_ms":100}`)
	source := &fakeLineSource{lines: [][]byte{ready, []byte(validIdentity)}}
	got, err := ReadIdentity(context.Background(), source, "fresh-token-123456", time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if got.DeviceID != "esp32-test" {
		t.Fatalf("unexpected identity: %#v", got)
	}
}

func TestReadIdentityRejectsChallengeMismatch(t *testing.T) {
	source := &fakeLineSource{lines: [][]byte{[]byte(validIdentity)}}
	_, err := ReadIdentity(context.Background(), source, "another-fresh-token", time.Second)
	if !errors.Is(err, ErrChallengeMismatch) {
		t.Fatalf("got %v, want challenge mismatch", err)
	}
}

func TestReadIdentityHonorsTimeout(t *testing.T) {
	start := time.Now()
	_, err := ReadIdentity(context.Background(), blockingLineSource{}, "fresh-token-123456", 20*time.Millisecond)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("got %v, want deadline exceeded", err)
	}
	if time.Since(start) > time.Second {
		t.Fatal("timeout was not bounded")
	}
}

func TestReadIdentityHonorsCallerCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := ReadIdentity(ctx, blockingLineSource{}, "fresh-token-123456", time.Second)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("got %v, want canceled", err)
	}
}

func TestReadIdentityRejectsInvalidArguments(t *testing.T) {
	if _, err := ReadIdentity(context.Background(), nil, "fresh-token-123456", time.Second); !errors.Is(err, ErrNoLineSource) {
		t.Fatalf("nil source: got %v", err)
	}
	if _, err := ReadIdentity(context.Background(), &fakeLineSource{}, "fresh-token-123456", 0); !errors.Is(err, ErrInvalidTimeout) {
		t.Fatalf("zero timeout: got %v", err)
	}
	if _, err := ReadIdentity(context.Background(), &fakeLineSource{}, "", time.Second); !errors.Is(err, ErrInvalidChallenge) {
		t.Fatalf("empty challenge: got %v", err)
	}
}

func TestReadIdentityReturnsSourceFailure(t *testing.T) {
	source := &fakeLineSource{err: io.EOF}
	_, err := ReadIdentity(context.Background(), source, "fresh-token-123456", time.Second)
	if !errors.Is(err, io.EOF) {
		t.Fatalf("got %v, want EOF", err)
	}
}
