package handshake

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"sync"
	"testing"
	"time"
)

// simulatedFirmware models the protocol-v1 firmware endpoint without opening a
// host serial port or touching hardware. It responds only to an identify line.
type simulatedFirmware struct {
	mu sync.Mutex
	mode string
	closed bool
	response []byte
}

func (s *simulatedFirmware) WriteLine(ctx context.Context, line []byte) error {
	if err := ctx.Err(); err != nil { return err }
	var request identifyRequest
	if err := json.Unmarshal(line, &request); err != nil { return err }
	if request.Protocol != ProtocolName || request.ProtocolVersion != ProtocolVersion || request.Command != "identify" {
		return errors.New("simulator received an invalid identify request")
	}
	if s.mode == "disconnect" { return io.ErrClosedPipe }
	if s.mode == "timeout" { return nil }
	challenge := request.Challenge
	if s.mode == "wrong-challenge" { challenge = "stale-response-token" }
	response := map[string]any{
		"protocol": ProtocolName, "protocol_version": ProtocolVersion,
		"event": "identity", "challenge": challenge,
		"firmware_version": "1.0.0", "project_id": "sim-project",
		"device_id": "sim-device", "uptime_ms": 4200,
	}
	encoded, err := json.Marshal(response)
	if err != nil { return err }
	if s.mode == "malformed" { encoded = []byte("{") }
	s.mu.Lock()
	s.response = encoded
	s.mu.Unlock()
	return nil
}

func (s *simulatedFirmware) ReadLine(ctx context.Context) ([]byte, error) {
	if err := ctx.Err(); err != nil { return nil, err }
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed { return nil, io.EOF }
	if s.response == nil {
		if s.mode == "timeout" { return nil, context.DeadlineExceeded }
		return nil, io.EOF
	}
	line := append([]byte(nil), s.response...)
	s.response = nil
	return line, nil
}

func (s *simulatedFirmware) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.closed = true
	return nil
}

func (s *simulatedFirmware) isClosed() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.closed
}

func TestSimulatedFirmwareHandshakeSucceedsWithoutHardware(t *testing.T) {
	sim := &simulatedFirmware{}
	got, err := RequestFreshIdentity(context.Background(), sim, time.Second)
	if err != nil { t.Fatal(err) }
	if got.ProjectID != "sim-project" || got.DeviceID != "sim-device" || got.FirmwareVersion != "1.0.0" {
		t.Fatalf("unexpected simulated identity: %+v", got)
	}
	if !sim.isClosed() { t.Fatal("simulator source was not closed") }
}

func TestSimulatedFirmwareHandshakeRejectsWrongChallenge(t *testing.T) {
	sim := &simulatedFirmware{mode: "wrong-challenge"}
	_, err := RequestFreshIdentity(context.Background(), sim, time.Second)
	if !errors.Is(err, ErrChallengeMismatch) { t.Fatalf("got %v, want challenge mismatch", err) }
	if !sim.isClosed() { t.Fatal("simulator source was not closed after mismatch") }
}

func TestSimulatedFirmwareHandshakeRejectsMalformedResponse(t *testing.T) {
	sim := &simulatedFirmware{mode: "malformed"}
	if _, err := RequestFreshIdentity(context.Background(), sim, time.Second); err == nil {
		t.Fatal("expected malformed simulator response to fail")
	}
	if !sim.isClosed() { t.Fatal("simulator source was not closed after malformed response") }
}

func TestSimulatedFirmwareHandshakeReportsDisconnect(t *testing.T) {
	sim := &simulatedFirmware{mode: "disconnect"}
	_, err := RequestFreshIdentity(context.Background(), sim, time.Second)
	if !errors.Is(err, ErrRequestWrite) || !errors.Is(err, io.ErrClosedPipe) {
		t.Fatalf("got %v, want wrapped disconnect/write error", err)
	}
	if !sim.isClosed() { t.Fatal("simulator source was not closed after disconnect") }
}

func TestSimulatedFirmwareHandshakeTimesOut(t *testing.T) {
	sim := &simulatedFirmware{mode: "timeout"}
	_, err := RequestFreshIdentity(context.Background(), sim, 20*time.Millisecond)
	if !errors.Is(err, context.DeadlineExceeded) { t.Fatalf("got %v, want deadline exceeded", err) }
	if !sim.isClosed() { t.Fatal("simulator source was not closed after timeout") }
}
