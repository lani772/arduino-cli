package firmware

import (
	"fmt"
	"strings"
)

const ClassicESP32DevModuleFQBN = "esp32:esp32:esp32"

type ESP32TargetAssessment struct {
	BoardFQBN string   `json:"board_fqbn"`
	Profile string     `json:"profile"`
	KnownProfile bool  `json:"known_profile"`
	Valid bool         `json:"valid"`
	Errors []string    `json:"errors"`
	Warnings []string  `json:"warnings"`
	Notes []string     `json:"notes"`
}

// AssessESP32Target checks only documented, conservative constraints for the
// known classic ESP32 Dev Module profile. Unknown ESP32 FQBNs are not treated
// as equivalent: they receive basic checks and an explicit verification warning.
func AssessESP32Target(fqbn string, statusGPIO int, lampGPIOs []int) ESP32TargetAssessment {
	a := ESP32TargetAssessment{
		BoardFQBN: fqbn,
		Profile: "unverified ESP32 variant",
		Valid: true,
		Errors: []string{},
		Warnings: []string{},
		Notes: []string{},
	}
	if strings.TrimSpace(fqbn) == "" {
		a.Valid = false
		a.Errors = append(a.Errors, "board_fqbn is required")
		return a
	}
	if !strings.HasPrefix(fqbn, "esp32:") {
		a.Valid = false
		a.Errors = append(a.Errors, "FQBN is not in the esp32 package family")
		return a
	}
	if fqbn == ClassicESP32DevModuleFQBN {
		a.KnownProfile = true
		a.Profile = "classic ESP32 Dev Module"
	} else {
		a.Warnings = append(a.Warnings, "No exact pin profile is registered for this FQBN; verify the exact board schematic and variant before generation or flashing.")
	}
	checkOutput := func(label string, gpio int) {
		if gpio < 0 || gpio > 39 {
			a.Valid = false
			a.Errors = append(a.Errors, fmt.Sprintf("%s GPIO %d is outside the supported classic ESP32 GPIO range 0-39", label, gpio))
			return
		}
		if a.KnownProfile && gpio >= 6 && gpio <= 11 {
			a.Valid = false
			a.Errors = append(a.Errors, fmt.Sprintf("%s GPIO %d is reserved for flash on the classic ESP32 profile", label, gpio))
		}
		if a.KnownProfile && gpio >= 34 && gpio <= 39 {
			a.Valid = false
			a.Errors = append(a.Errors, fmt.Sprintf("%s GPIO %d is input-only and cannot drive a lamp/status output on the classic ESP32 profile", label, gpio))
		}
		if a.KnownProfile && (gpio == 0 || gpio == 2 || gpio == 5 || gpio == 12 || gpio == 15) {
			a.Warnings = append(a.Warnings, fmt.Sprintf("%s GPIO %d is a boot-strapping pin; attached circuitry may prevent reliable startup", label, gpio))
		}
		if a.KnownProfile && (gpio == 1 || gpio == 3) {
			a.Warnings = append(a.Warnings, fmt.Sprintf("%s GPIO %d is commonly used for UART0; reassignment may interfere with serial logging or upload", label, gpio))
		}
	}
	checkOutput("status", statusGPIO)
	seen := map[int]string{}
	if statusGPIO >= 0 && statusGPIO <= 39 {
		seen[statusGPIO] = "status"
	}
	for i, gpio := range lampGPIOs {
		label := fmt.Sprintf("lamp[%d]", i)
		checkOutput(label, gpio)
		if previous, ok := seen[gpio]; ok {
			a.Valid = false
			a.Errors = append(a.Errors, fmt.Sprintf("%s GPIO %d duplicates %s assignment", label, gpio, previous))
		} else if gpio >= 0 && gpio <= 39 {
			seen[gpio] = label
		}
	}
	if !a.KnownProfile {
		a.Notes = append(a.Notes, "Basic GPIO range and duplicate checks do not prove board compatibility.")
	}
	return a
}
