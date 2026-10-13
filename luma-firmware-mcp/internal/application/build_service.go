package application

import (
 "context"
 "fmt"
 "strings"
 "github.com/lani772/arduino-cli/luma-firmware-mcp/internal/domain/build"
 "github.com/lani772/arduino-cli/luma-firmware-mcp/internal/domain/project"
)

type BuildService struct{ CLI ArduinoCLI }

func (s BuildService) Compile(ctx context.Context,p project.Project,fqbn string)(build.Result,error){
 if err:=p.Validate();err!=nil{return build.Result{},err}
 if strings.TrimSpace(fqbn)==""{return build.Result{},fmt.Errorf("board FQBN is required")}
 if s.CLI==nil{return build.Result{},fmt.Errorf("arduino CLI is required")}
 result,err:=s.CLI.Compile(ctx,p.Source,fqbn)
 if result.ID==""{result.ID=p.ID+"-build"}
 return result,err
}
