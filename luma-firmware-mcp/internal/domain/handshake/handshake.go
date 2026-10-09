package handshake

import (
	"bytes"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
)

const (
	ProtocolName      = "luma.runtime"
	ProtocolVersion   = 1
	MaxRecordBytes    = 4096
	MaxChallengeBytes = 256
)

var (
	ErrRecordTooLarge    = errors.New("handshake record exceeds maximum size")
	ErrInvalidChallenge  = errors.New("expected challenge must be non-empty and at most 256 bytes")
	ErrWrongProtocol     = errors.New("unsupported handshake protocol")
	ErrWrongVersion      = errors.New("unsupported handshake protocol version")
	ErrWrongEvent        = errors.New("record is not an identity response")
	ErrChallengeMismatch = errors.New("identity response challenge does not match")
)

type Identity struct {
	FirmwareVersion string `json:"firmware_version"`
	ProjectID       string `json:"project_id"`
	DeviceID        string `json:"device_id"`
	UptimeMS        int64  `json:"uptime_ms"`
}

type record struct {
	Protocol        string `json:"protocol"`
	ProtocolVersion int    `json:"protocol_version"`
	Event           string `json:"event"`
	Challenge       string `json:"challenge"`
	FirmwareVersion string `json:"firmware_version"`
	ProjectID       string `json:"project_id"`
	DeviceID        string `json:"device_id"`
	UptimeMS        *int64 `json:"uptime_ms"`
}

func validateExpectedChallenge(challenge string) error {
	if strings.TrimSpace(challenge) == "" || challenge != strings.TrimSpace(challenge) || len(challenge) > MaxChallengeBytes {
		return ErrInvalidChallenge
	}
	return nil
}

// ParseIdentity validates a single protocol-v1 identity record. It performs no I/O.
func ParseIdentity(line []byte, expectedChallenge string) (Identity, error) {
	var zero Identity
	if len(line) > MaxRecordBytes {
		return zero, ErrRecordTooLarge
	}
	if err := validateExpectedChallenge(expectedChallenge); err != nil {
		return zero, err
	}
	line = bytes.TrimSuffix(line, []byte("\n"))
	line = bytes.TrimSuffix(line, []byte("\r"))
	if len(bytes.TrimSpace(line)) == 0 {
		return zero, errors.New("empty handshake record")
	}
	dec := json.NewDecoder(bytes.NewReader(line))
	dec.DisallowUnknownFields()
	var msg record
	if err := dec.Decode(&msg); err != nil {
		return zero, fmt.Errorf("decode handshake record: %w", err)
	}
	var extra any
	if err := dec.Decode(&extra); !errors.Is(err, io.EOF) {
		return zero, errors.New("handshake input must contain exactly one JSON record")
	}
	if msg.Protocol != ProtocolName {
		return zero, ErrWrongProtocol
	}
	if msg.ProtocolVersion != ProtocolVersion {
		return zero, ErrWrongVersion
	}
	if msg.Event != "identity" {
		return zero, ErrWrongEvent
	}
	if msg.Challenge == "" || subtle.ConstantTimeCompare([]byte(msg.Challenge), []byte(expectedChallenge)) != 1 {
		return zero, ErrChallengeMismatch
	}
	if err := required("firmware_version", msg.FirmwareVersion, 128); err != nil {
		return zero, err
	}
	if err := required("project_id", msg.ProjectID, 128); err != nil {
		return zero, err
	}
	if err := required("device_id", msg.DeviceID, 128); err != nil {
		return zero, err
	}
	if msg.UptimeMS == nil || *msg.UptimeMS < 0 {
		return zero, errors.New("uptime_ms must be a non-negative integer")
	}
	return Identity{FirmwareVersion: msg.FirmwareVersion, ProjectID: msg.ProjectID, DeviceID: msg.DeviceID, UptimeMS: *msg.UptimeMS}, nil
}

func IsReadyRecord(line []byte) error {
	if len(line) > MaxRecordBytes {
		return ErrRecordTooLarge
	}
	dec := json.NewDecoder(bytes.NewReader(bytes.TrimSpace(line)))
	dec.DisallowUnknownFields()
	var msg record
	if err := dec.Decode(&msg); err != nil {
		return fmt.Errorf("decode readiness record: %w", err)
	}
	var extra any
	if err := dec.Decode(&extra); !errors.Is(err, io.EOF) {
		return errors.New("readiness input must contain exactly one JSON record")
	}
	if msg.Protocol != ProtocolName {
		return ErrWrongProtocol
	}
	if msg.ProtocolVersion != ProtocolVersion {
		return ErrWrongVersion
	}
	if msg.Event != "ready" {
		return ErrWrongEvent
	}
	if err := required("firmware_version", msg.FirmwareVersion, 128); err != nil {
		return err
	}
	if err := required("project_id", msg.ProjectID, 128); err != nil {
		return err
	}
	if err := required("device_id", msg.DeviceID, 128); err != nil {
		return err
	}
	if msg.UptimeMS == nil || *msg.UptimeMS < 0 {
		return errors.New("uptime_ms must be a non-negative integer")
	}
	return nil
}

func required(name, value string, max int) error {
	if strings.TrimSpace(value) == "" {
		return fmt.Errorf("%s is required", name)
	}
	if len(value) > max {
		return fmt.Errorf("%s exceeds %d bytes", name, max)
	}
	return nil
}
