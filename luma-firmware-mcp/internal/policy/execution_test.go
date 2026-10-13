package policy

import "testing"

func TestDefaultPolicyIsSafe(t *testing.T) {
 p := DefaultExecutionPolicy()
 if !p.AllowsExecutable("arduino-cli") { t.Fatal("arduino-cli should be allowed") }
 if p.AllowsExecutable("sh") || p.AllowsExecutable("bash") { t.Fatal("shells must not be allowed") }
 if p.AllowToolInstallation || p.AllowDeviceFlash { t.Fatal("installation and flashing must be disabled") }
}
