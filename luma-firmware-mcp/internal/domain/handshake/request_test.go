package handshake

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"testing"
	"time"
)

type fakeDuplexSource struct {
	fakeLineSource
	writes [][]byte
	writeErr error
}

func (f *fakeDuplexSource) WriteLine(ctx context.Context, line []byte) error {
	if err := ctx.Err(); err != nil { return err }
	f.writes = append(f.writes, append([]byte(nil), line...))
	return f.writeErr
}

func TestEncodeIdentifyRequestHasStrictProtocolFields(t *testing.T) {
	got, err := EncodeIdentifyRequest("fresh-token-123456")
	if err != nil { t.Fatal(err) }
	var fields map[string]any
	if err := json.Unmarshal(got, &fields); err != nil { t.Fatal(err) }
	if fields["protocol"] != ProtocolName || fields["protocol_version"] != float64(ProtocolVersion) ||
		fields["command"] != "identify" || fields["challenge"] != "fresh-token-123456" {
		t.Fatalf("unexpected request fields: %#v", fields)
	}
	if len(fields) != 4 { t.Fatalf("unexpected extra fields: %#v", fields) }
	if len(got) > MaxRecordBytes { t.Fatalf("request exceeds record limit: %d", len(got)) }
}

func TestEncodeIdentifyRequestRejectsInvalidChallenge(t *testing.T) {
	for _, challenge := range []string{"", " padded", "padded ", string(make([]byte, MaxChallengeBytes+1))} {
		if _, err := EncodeIdentifyRequest(challenge); !errors.Is(err, ErrInvalidChallenge) {
			t.Fatalf("challenge length %d: got %v", len(challenge), err)
		}
	}
}

func TestRequestIdentityWritesChallengeThenAcceptsMatchingResponse(t *testing.T) {
	challenge := "fresh-token-123456"
	response := []byte(`{"protocol":"luma.runtime","protocol_version":1,"event":"identity","challenge":"fresh-token-123456","firmware_version":"1.0.0","project_id":"luma-test","device_id":"esp32-test","uptime_ms":2345}`)
	source := &fakeDuplexSource{fakeLineSource: fakeLineSource{lines: [][]byte{response}}}
	got, err := RequestIdentity(context.Background(), source, challenge, time.Second)
	if err != nil { t.Fatal(err) }
	if got.DeviceID != "esp32-test" { t.Fatalf("unexpected identity: %#v", got) }
	if len(source.writes) != 1 { t.Fatalf("write count = %d, want 1", len(source.writes)) }
	want, err := EncodeIdentifyRequest(challenge)
	if err != nil { t.Fatal(err) }
	want = append(want, '\n')
	if string(source.writes[0]) != string(want) { t.Fatalf("request bytes = %q, want %q", source.writes[0], want) }
	if source.closed() != 1 { t.Fatalf("close count = %d, want 1", source.closed()) }
}

func TestRequestIdentityRejectsMismatchedResponse(t *testing.T) {
	source := &fakeDuplexSource{fakeLineSource: fakeLineSource{lines: [][]byte{[]byte(validIdentity)}}}
	_, err := RequestIdentity(context.Background(), source, "different-fresh-token", time.Second)
	if !errors.Is(err, ErrChallengeMismatch) { t.Fatalf("got %v, want challenge mismatch", err) }
	if source.closed() != 1 { t.Fatalf("close count = %d, want 1", source.closed()) }
}

func TestRequestIdentityClosesSourceOnWriteFailure(t *testing.T) {
	writeErr := errors.New("write failed")
	source := &fakeDuplexSource{fakeLineSource: fakeLineSource{err: io.EOF}, writeErr: writeErr}
	_, err := RequestIdentity(context.Background(), source, "fresh-token-123456", time.Second)
	if !errors.Is(err, ErrRequestWrite) || !errors.Is(err, writeErr) { t.Fatalf("got %v, want wrapped write error", err) }
	if source.closed() != 1 { t.Fatalf("close count = %d, want 1", source.closed()) }
}

func TestRequestIdentityRejectsInvalidArgumentsWithoutWritingOrClosing(t *testing.T) {
	source := &fakeDuplexSource{}
	if _, err := RequestIdentity(nil, source, "fresh-token-123456", time.Second); !errors.Is(err, ErrNoContext) {
		t.Fatalf("nil context: got %v", err)
	}
	if _, err := RequestIdentity(context.Background(), source, "", time.Second); !errors.Is(err, ErrInvalidChallenge) {
		t.Fatalf("empty challenge: got %v", err)
	}
	if len(source.writes) != 0 || source.closed() != 0 { t.Fatal("invalid arguments must not write or close source") }
}

func TestRequestIdentityRejectsMalformedResponseAndCloses(t *testing.T) {
	source := &fakeDuplexSource{fakeLineSource: fakeLineSource{lines: [][]byte{[]byte("{")}}}
	if _, err := RequestIdentity(context.Background(), source, "fresh-token-123456", time.Second); err == nil {
		t.Fatal("expected malformed response to be rejected")
	}
	if source.closed() != 1 {
		t.Fatalf("close count = %d, want 1", source.closed())
	}
}

func TestRequestIdentityDeadlineCoversWriteAndRead(t *testing.T) {
	source := &fakeDuplexSource{fakeLineSource: fakeLineSource{err: context.DeadlineExceeded}}
	_, err := RequestIdentity(context.Background(), source, "fresh-token-123456", 20*time.Millisecond)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("got %v, want deadline exceeded", err)
	}
	if source.closed() != 1 {
		t.Fatalf("close count = %d, want 1", source.closed())
	}
}
