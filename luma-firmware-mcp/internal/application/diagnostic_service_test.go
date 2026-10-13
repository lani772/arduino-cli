package application

import "testing"

func TestCompilerDiagnosticParser(t *testing.T) {
 out:=`src/main.cpp:42:7: error: 'foo' was not declared in this scope
src/main.cpp:43:2: warning: unused variable 'x'
note.cpp:9: note: informational note
irrelevant output`
 got:=CompilerDiagnosticParser{}.Parse(out)
 if len(got)!=3 {t.Fatalf("expected 3 diagnostics, got %d: %+v",len(got),got)}
 if got[0].File!="src/main.cpp"||got[0].Line!=42||got[0].Column!=7||!got[0].IsError(){t.Fatalf("unexpected error diagnostic: %+v",got[0])}
 if got[1].Severity.String()!="warning" {t.Fatalf("unexpected warning: %+v",got[1])}
 if got[2].Severity.String()!="info" {t.Fatalf("unexpected note: %+v",got[2])}
}
