package handshake

import (
	"context"
	"errors"
	"io"
	"sync"
	"testing"
	"time"
)

type fakeLineSource struct {
	lines      [][]byte
	index      int
	err        error
	closeCount int
	mu         sync.Mutex
}

func (f *fakeLineSource) ReadLine(ctx context.Context) ([]byte, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	f.mu.Lock()
	defer f.mu.Unlock()
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

func (f *fakeLineSource) Close() error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.closeCount++
	return nil
}

func (f *fakeLineSource) closed() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.closeCount
}

type blockingLineSource struct {
	once     sync.Once
	closedCh chan struct{}
}

func newBlockingLineSource() *blockingLineSource {
	return &blockingLineSource{closedCh: make(chan struct{})}
}

func (b *blockingLineSource) ReadLine(ctx context.Context) ([]byte, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-b.closedCh:
		return nil, io.ErrClosedPipe
	}
}

func (b *blockingLineSource) Close() error {
	b.once.Do(func() { close(b.closedCh) })
	return nil
}

func TestReadIdentitySkipsReadyThenAcceptsIdentityAndCloses(t *testing.T) {
	ready := []byte(`{"protocol":"luma.runtime","protocol_version":1,"event":"ready","firmware_version":"1.0.0","project_id":"luma-test","device_id":"esp32-test","uptime_ms":100}`)
	source := &fakeLineSource{lines: [][]byte{ready, []byte(validIdentity)}}
	got, err := ReadIdentity(context.Background(), source, "fresh-token-123456", time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if got.DeviceID != "esp32-test" {
		t.Fatalf("unexpected identity: %#v", got)
	}
	if source.closed() != 1 {
		t.Fatalf("source close count = %d, want 1", source.closed())
	}
}

func TestReadIdentityRejectsChallengeMismatchAndCloses(t *testing.T) {
	source := &fakeLineSource{lines: [][]byte{[]byte(validIdentity)}}
	_, err := ReadIdentity(context.Background(), source, "another-fresh-token", time.Second)
	if !errors.Is(err, ErrChallengeMismatch) {
		t.Fatalf("got %v, want challenge mismatch", err)
	}
	if source.closed() != 1 {
		t.Fatalf("source close count = %d, want 1", source.closed())
	}
}

func TestReadIdentityHonorsTimeoutAndClosesSource(t *testing.T) {
	source := newBlockingLineSource()
	start := time.Now()
	_, err := ReadIdentity(context.Background(), source, "fresh-token-123456", 20*time.Millisecond)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("got %v, want deadline exceeded", err)
	}
	if time.Since(start) > time.Second {
		t.Fatal("timeout was not bounded")
	}
	select {
	case <-source.closedCh:
	case <-time.After(time.Second):
		t.Fatal("source was not closed after timeout")
	}
}

func TestReadIdentityHonorsCallerCancellationAndCloses(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	source := newBlockingLineSource()
	cancel()
	_, err := ReadIdentity(ctx, source, "fresh-token-123456", time.Second)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("got %v, want canceled", err)
	}
	select {
	case <-source.closedCh:
	case <-time.After(time.Second):
		t.Fatal("source was not closed after cancellation")
	}
}

func TestReadIdentityRejectsInvalidArguments(t *testing.T) {
	if _, err := ReadIdentity(context.Background(), nil, "fresh-token-123456", time.Second); !errors.Is(err, ErrNoLineSource) {
		t.Fatalf("nil source: got %v", err)
	}
	source := &fakeLineSource{}
	if _, err := ReadIdentity(context.Background(), source, "fresh-token-123456", 0); !errors.Is(err, ErrInvalidTimeout) {
		t.Fatalf("zero timeout: got %v", err)
	}
	if source.closed() != 0 {
		t.Fatal("source should not be closed for invalid timeout")
	}
	if _, err := ReadIdentity(context.Background(), source, "", time.Second); !errors.Is(err, ErrInvalidChallenge) {
		t.Fatalf("empty challenge: got %v", err)
	}
	if source.closed() != 0 {
		t.Fatal("source should not be closed for invalid challenge")
	}
}

func TestReadIdentityReturnsSourceFailureAndCloses(t *testing.T) {
	source := &fakeLineSource{err: io.EOF}
	_, err := ReadIdentity(context.Background(), source, "fresh-token-123456", time.Second)
	if !errors.Is(err, io.EOF) {
		t.Fatalf("got %v, want EOF", err)
	}
	if source.closed() != 1 {
		t.Fatalf("source close count = %d, want 1", source.closed())
	}
}
