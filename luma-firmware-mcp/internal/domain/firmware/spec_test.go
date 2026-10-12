package firmware
import "testing"
func TestSpecificationRejectsDuplicateGPIO(t *testing.T){s:=Specification{ProjectID:"p",MicrocontrollerName:"m",Target:Target{BoardFQBN:"esp32:esp32:esp32"},FirmwareVersion:"0.1.0",StatusGPIO:13,Lamps:[]Lamp{{ID:"1",Name:"A",GPIO:14},{ID:"2",Name:"B",GPIO:14}}};if err:=s.Validate();err==nil{t.Fatal("expected duplicate GPIO validation error")}}
func TestSpecificationDefaultsStatusGPIO(t *testing.T){s:=Specification{ProjectID:"p",MicrocontrollerName:"m",Target:Target{BoardFQBN:"esp32:esp32:esp32"},FirmwareVersion:"0.1.0",Lamps:[]Lamp{{ID:"1",Name:"A",GPIO:14}}};s=s.WithDefaults();if err:=s.Validate();err!=nil{t.Fatal(err)};if s.StatusGPIO!=DefaultStatusGPIO{t.Fatalf("expected status GPIO %d, got %d",DefaultStatusGPIO,s.StatusGPIO)}}

func TestSpecificationRejectsInputOnlyGPIOOnClassicESP32(t *testing.T) {
	for _, gpio := range []int{34, 35, 36, 39} {
		s := Specification{ProjectID:"p", MicrocontrollerName:"m", Target:Target{BoardFQBN:ClassicESP32DevModuleFQBN}, FirmwareVersion:"0.1.0", StatusGPIO:13, Lamps:[]Lamp{{ID:"1", Name:"A", GPIO:gpio}}}
		if err := s.Validate(); err == nil {
			t.Fatalf("expected GPIO %d to be rejected as input-only", gpio)
		}
	}
}
func TestSpecificationRejectsInputOnlyStatusGPIOOnClassicESP32(t *testing.T) {
	s := Specification{ProjectID:"p", MicrocontrollerName:"m", Target:Target{BoardFQBN:ClassicESP32DevModuleFQBN}, FirmwareVersion:"0.1.0", StatusGPIO:34, Lamps:[]Lamp{{ID:"1", Name:"A", GPIO:14}}}
	if err := s.Validate(); err == nil {
		t.Fatal("expected input-only status GPIO to be rejected")
	}
}
func TestSpecificationDoesNotAssumeClassicPinoutForUnknownESP32(t *testing.T) {
	s := Specification{ProjectID:"p", MicrocontrollerName:"m", Target:Target{BoardFQBN:"esp32:esp32:esp32s3"}, FirmwareVersion:"0.1.0", StatusGPIO:13, Lamps:[]Lamp{{ID:"1", Name:"A", GPIO:34}}}
	if err := s.Validate(); err != nil {
		t.Fatalf("unknown variant should not inherit classic pinout rules: %v", err)
	}
}
