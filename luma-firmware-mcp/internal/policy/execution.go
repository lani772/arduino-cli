package policy

type ExecutionPolicy struct {
 AllowProcessExecution bool
 AllowToolInstallation bool
 AllowDeviceFlash bool
 AllowedExecutables []string
}

func DefaultExecutionPolicy() ExecutionPolicy {
 return ExecutionPolicy{AllowProcessExecution:true, AllowToolInstallation:false, AllowDeviceFlash:false, AllowedExecutables:[]string{"arduino-cli"}}
}

func (p ExecutionPolicy) AllowsExecutable(name string) bool {
 if !p.AllowProcessExecution { return false }
 for _, allowed := range p.AllowedExecutables { if allowed == name { return true } }
 return false
}
