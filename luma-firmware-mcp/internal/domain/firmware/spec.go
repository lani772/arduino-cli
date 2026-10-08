package firmware

import (
 "fmt"
 "github.com/lani772/arduino-cli/luma-firmware-mcp/internal/domain/errors"
)

const DefaultStatusGPIO = 13
type Lamp struct { ID string; Name string; Room string; GPIO int }
type Specification struct {
 ProjectID string
 MicrocontrollerName string
 Target Target
 FirmwareVersion string
 Lamps []Lamp
 StatusGPIO int
 OfflineSchedules bool
 MQTTEnabled bool
}
func (s Specification) WithDefaults() Specification { if s.StatusGPIO == 0 { s.StatusGPIO = DefaultStatusGPIO }; return s }
func (s Specification) Validate() error {
 if s.ProjectID == "" { return errors.Invalid("project_id","project id is required") }
 if s.MicrocontrollerName == "" { return errors.Invalid("microcontroller_name","microcontroller name is required") }
 if err:=s.Target.Validate(); err!=nil{return err}
 if s.FirmwareVersion == "" { return errors.Invalid("firmware_version","firmware version is required") }
 if len(s.Lamps)==0{return errors.Invalid("lamps","at least one lamp is required")}
 if s.StatusGPIO<0||s.StatusGPIO>39{return errors.Invalid("status_gpio","status GPIO must be between 0 and 39")}
 seen:=map[int]bool{s.StatusGPIO:true}
 for i,l:=range s.Lamps {
  if l.ID==""{return errors.Invalid(fmt.Sprintf("lamps[%d].id",i),"lamp id is required")}
  if l.Name==""{return errors.Invalid(fmt.Sprintf("lamps[%d].name",i),"lamp name is required")}
  if l.GPIO<0||l.GPIO>39{return errors.Invalid(fmt.Sprintf("lamps[%d].gpio",i),"GPIO must be between 0 and 39")}
  if seen[l.GPIO]{return errors.Invalid(fmt.Sprintf("lamps[%d].gpio",i),"GPIO is already reserved or assigned")}
  seen[l.GPIO]=true
  if l.GPIO>=6&&l.GPIO<=11{return errors.Invalid(fmt.Sprintf("lamps[%d].gpio",i),"GPIO 6-11 is reserved on common ESP32 modules")}
 }
 return nil
}
type SourceFile struct { Path string; Content string }
type GeneratedProject struct { ProjectID string; Files []SourceFile }
func (p GeneratedProject) Validate() error {
 if p.ProjectID==""{return errors.Invalid("project_id","project id is required")}
 if len(p.Files)==0{return errors.Invalid("files","at least one generated source file is required")}
 for i,f:=range p.Files { if f.Path==""{return errors.Invalid(fmt.Sprintf("files[%d].path",i),"source file path is required")}; if f.Content==""{return errors.Invalid(fmt.Sprintf("files[%d].content",i),"source file content is required")} }
 return nil
}
