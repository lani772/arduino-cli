package handshake

import (
	"context"
	"errors"
	"io"
	"testing"
	"time"
)

type fakeTransport struct {
	source   LineSource
	err      error
	calls    int
	lastPort string
}

func (f *fakeTransport) Open(ctx context.Context, port string) (LineSource, error) {
	f.calls++
	f.lastPort = port
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return f.source, f.err
}

func TestOpenSourceRequiresConsentAndExactPortBeforeOpening(t *testing.T) {
	transport := &fakeTransport{source: &fakeLineSource{}}
	if _, err := OpenSource(context.Background(), transport, "COM4", false); !errors.Is(err, ErrNoOperatorConsent) {
		t.Fatalf("got %v, want consent error", err)
	}
	for _, port := range []string{" ", " COM4", "COM4 "} {
		if _, err := OpenSource(context.Background(), transport, port, true); !errors.Is(err, ErrNoPort) {
			t.Fatalf("port %q: got %v, want selected-port error", port, err)
		}
	}
	if transport.calls != 0 {
		t.Fatalf("transport opened %d times before validation", transport.calls)
	}
}

func TestOpenSourceUsesOnlySelectedPortAndTransfersOwnership(t *testing.T) {
	source := &fakeLineSource{err: io.EOF}
	transport := &fakeTransport{source: source}
	got, err := OpenSource(context.Background(), transport, "/dev/ttyUSB2", true)
	if err != nil {
		t.Fatal(err)
	}
	if got != source {
		t.Fatal("returned source differs from transport source")
	}
	if transport.calls != 1 || transport.lastPort != "/dev/ttyUSB2" {
		t.Fatalf("unexpected open call: calls=%d port=%q", transport.calls, transport.lastPort)
	}
	_, err = ReadIdentity(context.Background(), got, "fresh-token-123456", time.Second)
	if !errors.Is(err, io.EOF) {
		t.Fatalf("got %v, want EOF", err)
	}
	if source.closed() != 1 {
		t.Fatalf("source close count = %d, want 1", source.closed())
	}
}

func TestOpenSourceRejectsCancelledContextWithoutOpening(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	transport := &fakeTransport{source: &fakeLineSource{}}
	if _, err := OpenSource(ctx, transport, "COM4", true); !errors.Is(err, context.Canceled) {
		t.Fatalf("got %v, want context canceled", err)
	}
	if transport.calls != 0 {
		t.Fatalf("transport opened %d times for cancelled context", transport.calls)
	}
}

func TestOpenSourceWrapsOpenFailure(t *testing.T) {
	openErr := errors.New("port busy")
	transport := &fakeTransport{err: openErr}
	if _, err := OpenSource(context.Background(), transport, "COM4", true); !errors.Is(err, ErrTransportOpen) || !errors.Is(err, openErr) {
		t.Fatalf("got %v, want wrapped transport and source errors", err)
	}
}

func TestOpenSourceRejectsNilSource(t *testing.T) {
	transport := &fakeTransport{}
	if _, err := OpenSource(context.Background(), transport, "COM4", true); !errors.Is(err, ErrTransportOpen) {
		t.Fatalf("got %v, want transport-open error", err)
	}
}

func TestOpenSourceRejectsInvalidArgumentsWithoutOpening(t *testing.T) {
	transport := &fakeTransport{source: &fakeLineSource{}}
	if _, err := OpenSource(nil, transport, "COM4", true); !errors.Is(err, ErrNoContext) {
		t.Fatalf("nil context: got %v", err)
	}
	if _, err := OpenSource(context.Background(), nil, "COM4", true); !errors.Is(err, ErrNoTransport) {
		t.Fatalf("nil transport: got %v", err)
	}
	if transport.calls != 0 {
		t.Fatalf("transport opened %d times for invalid arguments", transport.calls)
	}
}
