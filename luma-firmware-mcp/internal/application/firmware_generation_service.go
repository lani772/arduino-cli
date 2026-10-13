package application

import (
 "context"
 "encoding/json"
 "fmt"
 "strings"
 "github.com/lani772/arduino-cli/luma-firmware-mcp/internal/domain/firmware"
)
type FirmwareGenerationService struct { Provider AIProvider }
func (s FirmwareGenerationService) Generate(ctx context.Context,input firmware.Specification)(firmware.GeneratedProject,error){
 spec:=input.WithDefaults()
 if err:=spec.Validate();err!=nil{return firmware.GeneratedProject{},err}
 if s.Provider==nil{return firmware.GeneratedProject{},fmt.Errorf("AI provider is required")}
 instruction,err:=buildFirmwareInstruction(spec);if err!=nil{return firmware.GeneratedProject{},err}
 response,err:=s.Provider.Complete(ctx,AIRequest{Operation:"generate_firmware",Instruction:instruction});if err!=nil{return firmware.GeneratedProject{},err}
 content:=strings.TrimSpace(response.Content);if content==""{return firmware.GeneratedProject{},fmt.Errorf("AI returned empty firmware")}
 var envelope struct{ProjectID string;Files []firmware.SourceFile}
 if err:=json.Unmarshal([]byte(content),&envelope);err!=nil{return firmware.GeneratedProject{},fmt.Errorf("AI firmware response must be JSON: %w",err)}
 generated:=firmware.GeneratedProject{ProjectID:envelope.ProjectID,Files:envelope.Files};if generated.ProjectID==""{generated.ProjectID=spec.ProjectID}
 if err:=generated.Validate();err!=nil{return firmware.GeneratedProject{},err};return generated,nil
}
func buildFirmwareInstruction(spec firmware.Specification)(string,error){
 payload:=map[string]interface{}{"project_id":spec.ProjectID,"microcontroller_name":spec.MicrocontrollerName,"board_fqbn":spec.Target.BoardFQBN,"firmware_version":spec.FirmwareVersion,"status_gpio":spec.StatusGPIO,"lamps":spec.Lamps,"offline_schedules":spec.OfflineSchedules,"mqtt_enabled":spec.MQTTEnabled}
 b,err:=json.Marshal(payload);if err!=nil{return "",err}
 return "Generate only a LUMA ESP32 firmware project from this specification.\nHard requirements:\n- Preserve every requested lamp ID, name, room, and GPIO assignment.\n- Never invent, duplicate, or reassign GPIOs.\n- Never embed credentials, tokens, private keys, MQTT passwords, or other secrets.\n- Keep device identity and cloud authorization outside generated source unless explicitly provided as non-secret configuration.\n- Return JSON only with this shape: {project_id:..., files:[{path:..., content:...}]}.\n- Include a compilable .ino entrypoint.\n- Implement lamp outputs with safe startup and explicit HIGH/LOW state handling.\n- Use status GPIO 13 unless the specification explicitly provides another status GPIO.\n- Offline schedules must be represented only when enabled.\n- MQTT integration must use placeholders/interfaces when enabled; never fabricate credentials or broker secrets.\nSpecification:\n"+string(b),nil
}
