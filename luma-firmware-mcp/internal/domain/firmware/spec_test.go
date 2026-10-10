package firmware
import "testing"
func TestSpecificationRejectsDuplicateGPIO(t *testing.T){s:=Specification{ProjectID:"p",MicrocontrollerName:"m",Target:Target{BoardFQBN:"esp32:esp32:esp32"},FirmwareVersion:"0.1.0",StatusGPIO:13,Lamps:[]Lamp{{ID:"1",Name:"A",GPIO:14},{ID:"2",Name:"B",GPIO:14}}};if err:=s.Validate();err==nil{t.Fatal("expected duplicate GPIO validation error")}}
func TestSpecificationDefaultsStatusGPIO(t *testing.T){s:=Specification{ProjectID:"p",MicrocontrollerName:"m",Target:Target{BoardFQBN:"esp32:esp32:esp32"},FirmwareVersion:"0.1.0",Lamps:[]Lamp{{ID:"1",Name:"A",GPIO:14}}};s=s.WithDefaults();if err:=s.Validate();err!=nil{t.Fatal(err)};if s.StatusGPIO!=DefaultStatusGPIO{t.Fatalf("expected status GPIO %d, got %d",DefaultStatusGPIO,s.StatusGPIO)}}
