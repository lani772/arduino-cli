package repair

import "testing"

func TestApplyUpdateCreateDeleteAtomically(t *testing.T) {
 original := map[string]string{"main.ino": "old"}
 patch := Patch{ID:"p1", Description:"update firmware", Status:StatusProposed, Files:[]FileChange{
  {Path:"main.ino", Operation:OperationUpdate, Content:"new", ExpectedSHA:ContentSHA("old")},
  {Path:"include/config.h", Operation:OperationCreate, Content:"#pragma once"},
 }}
 got, err := Apply(original, patch)
 if err != nil { t.Fatal(err) }
 if got["main.ino"] != "new" || got["include/config.h"] != "#pragma once" { t.Fatalf("unexpected result: %#v", got) }
 if original["main.ino"] != "old" { t.Fatal("Apply mutated the original snapshot") }

 del := Patch{ID:"p2", Description:"remove config", Status:StatusProposed, Files:[]FileChange{
  {Path:"include/config.h", Operation:OperationDelete, ExpectedSHA:ContentSHA("#pragma once")},
 }}
 got, err = Apply(got, del)
 if err != nil { t.Fatal(err) }
 if _, ok := got["include/config.h"]; ok { t.Fatal("expected file to be deleted") }
}

func TestApplyRejectsUnsafePathsAndStaleHashes(t *testing.T) {
 cases := []Patch{
  {ID:"p", Description:"escape", Status:StatusProposed, Files:[]FileChange{{Path:"../outside.ino", Operation:OperationCreate, Content:"x"}}},
  {ID:"p", Description:"stale", Status:StatusProposed, Files:[]FileChange{{Path:"main.ino", Operation:OperationUpdate, Content:"new", ExpectedSHA:ContentSHA("not-current")}}},
 }
 for i, patch := range cases {
  _, err := Apply(map[string]string{"main.ino":"old"}, patch)
  if err == nil { t.Fatalf("case %d: expected error", i) }
 }
}

func TestApplyIsAtomicOnLaterFailure(t *testing.T) {
 source := map[string]string{"main.ino":"old"}
 patch := Patch{ID:"p", Description:"atomic", Status:StatusProposed, Files:[]FileChange{
  {Path:"new.ino", Operation:OperationCreate, Content:"new"},
  {Path:"main.ino", Operation:OperationUpdate, Content:"wrong", ExpectedSHA:ContentSHA("stale")},
 }}
 got, err := Apply(source, patch)
 if err == nil || got != nil { t.Fatalf("expected nil result and error, got=%v err=%v", got, err) }
 if source["main.ino"] != "old" { t.Fatal("original source changed after failed patch") }
}

func TestValidateAllowsLegacyDisplayDiff(t *testing.T) {
 patch := Patch{ID:"legacy", Description:"old proposal", Diff:"--- a/main.ino", Status:StatusProposed}
 if err := patch.Validate(); err != nil { t.Fatal(err) }
}
