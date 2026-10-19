package handshake

import (
	"errors"
	"fmt"
	"strings"
)

// ExpectedIdentity contains optional deployment values that the caller already
// trusts. Empty fields are not compared; values reported by firmware are never
// promoted to trusted configuration merely because they are well-formed.
type ExpectedIdentity struct {
	FirmwareVersion string
	ProjectID       string
	DeviceID        string
}

var (
	ErrIdentityMismatch = errors.New("runtime identity does not match expected configuration")
	ErrInvalidVersion   = errors.New("firmware version must be major.minor.patch")
)

// ValidateIdentity compares a parsed response with caller-provided expectations.
// It validates the reported version as numeric major.minor.patch and compares
// every non-empty expected field exactly. It does not authenticate the device.
func ValidateIdentity(got Identity, expected ExpectedIdentity) error {
	if err := validateVersion(got.FirmwareVersion); err != nil {
		return err
	}
	if err := required("project_id", got.ProjectID, 128); err != nil {
		return err
	}
	if err := required("device_id", got.DeviceID, 128); err != nil {
		return err
	}
	if expected.FirmwareVersion != "" {
		if err := validateVersion(expected.FirmwareVersion); err != nil {
			return fmt.Errorf("expected firmware version: %w", err)
		}
		if got.FirmwareVersion != expected.FirmwareVersion {
			return fmt.Errorf("%w: firmware_version", ErrIdentityMismatch)
		}
	}
	if expected.ProjectID != "" && got.ProjectID != expected.ProjectID {
		return fmt.Errorf("%w: project_id", ErrIdentityMismatch)
	}
	if expected.DeviceID != "" && got.DeviceID != expected.DeviceID {
		return fmt.Errorf("%w: device_id", ErrIdentityMismatch)
	}
	return nil
}

func validateVersion(value string) error {
	if value == "" || strings.TrimSpace(value) != value {
		return ErrInvalidVersion
	}
	parts := strings.Split(value, ".")
	if len(parts) != 3 {
		return ErrInvalidVersion
	}
	for _, part := range parts {
		if part == "" {
			return ErrInvalidVersion
		}
		for _, r := range part {
			if r < '0' || r > '9' {
				return ErrInvalidVersion
			}
		}
		if len(part) > 1 && part[0] == '0' {
			return ErrInvalidVersion
		}
	}
	return nil
}
