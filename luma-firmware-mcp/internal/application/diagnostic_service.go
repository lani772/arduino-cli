package application

import (
 "regexp"
 "strconv"
 "strings"
 "github.com/lani772/arduino-cli/luma-firmware-mcp/internal/domain/diagnostic"
)

type DiagnosticParser interface { Parse(string) []diagnostic.Diagnostic }

type CompilerDiagnosticParser struct{}

var compilerLine=regexp.MustCompile("^(.+?):([0-9]+):([0-9]+):\\s*(fatal error|error|warning|note):\\s*(.+)$")
var simpleLine=regexp.MustCompile("^(.+?):([0-9]+):\\s*(fatal error|error|warning|note):\\s*(.+)$")

func (CompilerDiagnosticParser) Parse(output string) []diagnostic.Diagnostic {
 var result []diagnostic.Diagnostic
 for _,line:=range strings.Split(output,"\n") {
  line=strings.TrimSpace(line); if line=="" {continue}
  if m:=compilerLine.FindStringSubmatch(line);m!=nil {result=append(result,makeDiagnostic(m[1],m[2],m[3],m[4],m[5]));continue}
  if m:=simpleLine.FindStringSubmatch(line);m!=nil {result=append(result,makeDiagnostic(m[1],m[2],"0",m[3],m[4]))}
 }
 return result
}
func makeDiagnostic(file,line,column,severity,message string) diagnostic.Diagnostic {
 s:=diagnostic.SeverityInfo
 switch strings.ToLower(severity) {case "error","fatal error":s=diagnostic.SeverityError;case "warning":s=diagnostic.SeverityWarning}
 return diagnostic.Diagnostic{Severity:s,Code:strings.ToLower(strings.ReplaceAll(severity," ","_")),Message:strings.TrimSpace(message),File:file,Line:atoi(line),Column:atoi(column)}
}
func atoi(v string) int {n,_:=strconv.Atoi(v);return n}
