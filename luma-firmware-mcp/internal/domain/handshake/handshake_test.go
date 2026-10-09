package handshake

import (
	"errors"
	"strings"
	"testing"
)

const validIdentity = `{"protocol":"luma.runtime","protocol_version":1,"event":"identity","challenge":"fresh-token-123456","firmware_version":"1.0.0","project_id":"luma-test","device_id":"esp32-test","uptime_ms":2345}`

func TestParseIdentityValidResponse(t *testing.T) {
	got, err := ParseIdentity([]byte(validIdentity+"\n"), "fresh-token-123456")
	if err != nil {
		t.Fatal(err)
	}
	if got.FirmwareVersion != "1.0.0" || got.ProjectID != "luma-test" || got.DeviceID != "esp32-test" || got.UptimeMS != 2345 {
		t.Fatalf("unexpected identity: %#v", got)
	}
}

func TestParseIdentityRejectsInvalidRecords(t *testing.T) {
	tests := []struct {
		name, raw, challenge string
		want error
	}{
		{"malformed JSON", "{", "fresh-token-123456", nil},
		{"wrong protocol", strings.Replace(validIdentity, "luma.runtime", "other.runtime", 1), "fresh-token-123456", ErrWrongProtocol},
		{"unsupported version", strings.Replace(validIdentity, `"protocol_version":1`, `"protocol_version":2`, 1), "fresh-token-123456", ErrWrongVersion},
		{"wrong event", strings.Replace(validIdentity, `"identity"`, `"ready"`, 1), "fresh-token-123456", ErrWrongEvent},
		{"challenge mismatch", validIdentity, "another-fresh-token", ErrChallengeMismatch},
		{"missing firmware version", strings.Replace(validIdentity, `"firmware_version":"1.0.0",`, "", 1), "fresh-token-123456", nil},
		{"empty project", strings.Replace(validIdentity, `"project_id":"luma-test"`, `"project_id":""`, 1), "fresh-token-123456", nil},
		{"empty device", strings.Replace(validIdentity, `"device_id":"esp32-test"`, `"device_id":""`, 1), "fresh-token-123456", nil},
		{"negative uptime", strings.Replace(validIdentity, `"uptime_ms":2345`, `"uptime_ms":-1`, 1), "fresh-token-123456", nil},
		{"string uptime", strings.Replace(validIdentity, `"uptime_ms":2345`, `"uptime_ms":"2345"`, 1), "fresh-token-123456", nil},
		{"extra JSON value", validIdentity + " {}", "fresh-token-123456", nil},
		{"unknown field", strings.TrimSuffix(validIdentity, "}") + `, "extra":true}`, "fresh-token-123456", nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseIdentity([]byte(tt.raw), tt.challenge)
			if err == nil {
				t.Fatal("expected rejection")
			}
			if tt.want != nil && !errors.Is(err, tt.want) {
				t.Fatalf("error %v, want %v", err, tt.want)
			}
		})
	}
}

func TestParseIdentityRejectsEmptyOrOversizedExpectedChallenge(t *testing.T) {
	for _, challenge := range []string{"", "  ", " fresh-token-123456", strings.Repeat("x", MaxChallengeBytes+1)} {
		if _, err := ParseIdentity([]byte(validIdentity), challenge); !errors.Is(err, ErrInvalidChallenge) {
			t.Fatalf("challenge length %d: got %v", len(challenge), err)
		}
	}
}

func TestParseIdentityRejectsOversizedRecord(t *testing.T) {
	if _, err := ParseIdentity([]byte(strings.Repeat("x", MaxRecordBytes+1)), "fresh-token-123456"); !errors.Is(err, ErrRecordTooLarge) {
		t.Fatalf("got %v", err)
	}
}

func TestReadyRecordIsNotIdentityResponse(t *testing.T) {
	ready := strings.Replace(validIdentity, `"identity"`, `"ready"`, 1)
	if err := IsReadyRecord([]byte(ready)); err != nil {
		t.Fatalf("valid ready record rejected: %v", err)
	}
	if _, err := ParseIdentity([]byte(ready), "fresh-token-123456"); !errors.Is(err, ErrWrongEvent) {
		t.Fatalf("ready record incorrectly accepted as identity: %v", err)
	}
}

func TestIsReadyRecordRejectsMalformedAndInvalidRecords(t *testing.T) {
	if err := IsReadyRecord([]byte("{")); err == nil {
		t.Fatal("expected malformed JSON rejection")
	}
	ready := strings.Replace(validIdentity, `"identity"`, `"ready"`, 1)
	if err := IsReadyRecord([]byte(strings.TrimSuffix(ready, "}") + `, "unexpected":true}`)); err == nil {
		t.Fatal("expected unknown readiness field rejection")
	}
	if err := IsReadyRecord([]byte(ready + " {}")); err == nil {
		t.Fatal("expected trailing JSON value rejection")
	}
	negative := strings.Replace(ready, `"uptime_ms":2345`, `"uptime_ms":-2`, 1)
	if err := IsReadyRecord([]byte(negative)); err == nil {
		t.Fatal("expected negative uptime rejection")
	}
}
