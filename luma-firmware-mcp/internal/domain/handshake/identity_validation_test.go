package handshake

import (
	"errors"
	"testing"
)

func TestValidateIdentityMatchesExpectedConfiguration(t *testing.T) {
	got := Identity{FirmwareVersion: "1.2.3", ProjectID: "project-a", DeviceID: "esp32-ABCDEF", UptimeMS: 42}
	if err := ValidateIdentity(got, ExpectedIdentity{FirmwareVersion: "1.2.3", ProjectID: "project-a", DeviceID: "esp32-ABCDEF"}); err != nil {
		t.Fatalf("valid identity rejected: %v", err)
	}
}

func TestValidateIdentityRejectsExpectedValueMismatches(t *testing.T) {
	got := Identity{FirmwareVersion: "1.2.3", ProjectID: "project-a", DeviceID: "esp32-ABCDEF"}
	tests := []struct {
		name string
		want ExpectedIdentity
	}{
		{"firmware version", ExpectedIdentity{FirmwareVersion: "1.2.4"}},
		{"project ID", ExpectedIdentity{ProjectID: "project-b"}},
		{"device ID", ExpectedIdentity{DeviceID: "esp32-OTHER"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ValidateIdentity(got, tt.want); !errors.Is(err, ErrIdentityMismatch) {
				t.Fatalf("got %v, want identity mismatch", err)
			}
		})
	}
}

func TestValidateIdentityRejectsInvalidVersion(t *testing.T) {
	for _, version := range []string{"", "1", "1.2", "v1.2.3", "1.02.3", "1.2.x", " 1.2.3"} {
		t.Run(version, func(t *testing.T) {
			got := Identity{FirmwareVersion: version, ProjectID: "project-a", DeviceID: "device-a"}
			if err := ValidateIdentity(got, ExpectedIdentity{}); !errors.Is(err, ErrInvalidVersion) {
				t.Fatalf("got %v, want invalid version", err)
			}
		})
	}
}

func TestValidateIdentityAllowsUnspecifiedExpectedFields(t *testing.T) {
	got := Identity{FirmwareVersion: "0.1.0", ProjectID: "reference", DeviceID: "esp32-123"}
	if err := ValidateIdentity(got, ExpectedIdentity{}); err != nil {
		t.Fatalf("well-formed self-reported identity rejected: %v", err)
	}
}
