package firmware

import (
	"strings"
	"testing"
)

func TestAssessESP32KnownProfileRejectsFlashAndInputOnlyGPIO(t *testing.T) {
	a := AssessESP32Target(ClassicESP32DevModuleFQBN, 13, []int{14, 34, 6})
	if a.Valid {
		t.Fatal("expected invalid GPIO assignments")
	}
	joined := strings.Join(a.Errors, " ")
	if !strings.Contains(joined, "input-only") || !strings.Contains(joined, "reserved for flash") {
		t.Fatalf("expected flash and input-only errors, got %v", a.Errors)
	}
}

func TestAssessESP32WarnsOnStrappingPin(t *testing.T) {
	a := AssessESP32Target(ClassicESP32DevModuleFQBN, 13, []int{12})
	if !a.Valid || len(a.Warnings) == 0 || !strings.Contains(strings.Join(a.Warnings, " "), "boot-strapping") {
		t.Fatalf("expected valid assessment with strapping warning, got %+v", a)
	}
}

func TestAssessESP32UnknownVariantRequiresVerification(t *testing.T) {
	a := AssessESP32Target("esp32:esp32:esp32s3", 13, []int{14})
	if !a.Valid || a.KnownProfile || len(a.Warnings) == 0 {
		t.Fatalf("expected unverified warning for unknown variant, got %+v", a)
	}
}

func TestAssessESP32RejectsNonESP32FQBNAndDuplicates(t *testing.T) {
	if a := AssessESP32Target("arduino:avr:uno", 13, []int{14}); a.Valid {
		t.Fatal("expected non-ESP32 FQBN to be rejected")
	}
	if a := AssessESP32Target(ClassicESP32DevModuleFQBN, 13, []int{13}); a.Valid {
		t.Fatal("expected duplicate status/lamp GPIO to be rejected")
	}
}
