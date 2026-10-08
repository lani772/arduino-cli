package ai

import (
 "context"
 "github.com/lani772/arduino-cli/luma-firmware-mcp/internal/application"
)

type FakeProvider struct {
 Response application.AIResponse
 Requests []application.AIRequest
 Err error
}
func(f *FakeProvider)Complete(_ context.Context,r application.AIRequest)(application.AIResponse,error){
 f.Requests=append(f.Requests,r);if f.Err!=nil{return application.AIResponse{},f.Err};return f.Response,nil
}
