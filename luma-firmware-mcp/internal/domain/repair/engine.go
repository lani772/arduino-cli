package repair

import (
 "crypto/sha256"
 "encoding/hex"
 "fmt"
 "path"
 "strings"
)

// ContentSHA returns the SHA-256 hex digest used by ExpectedSHA.
func ContentSHA(content string) string {
 sum := sha256.Sum256([]byte(content))
 return hex.EncodeToString(sum[:])
}

// Apply validates and applies all file changes to an in-memory source snapshot.
// It returns a new map only if every change succeeds; the input is never mutated.
func Apply(source map[string]string, patch Patch) (map[string]string, error) {
 if err := patch.Validate(); err != nil { return nil, err }
 if patch.Status != StatusProposed { return nil, fmt.Errorf("only proposed patches can be applied") }
 if len(patch.Files) == 0 { return nil, fmt.Errorf("patch contains no structured file changes") }

 next := make(map[string]string, len(source)+len(patch.Files))
 for name, content := range source {
  if err := validatePath(name); err != nil { return nil, fmt.Errorf("source snapshot: %w", err) }
  next[name] = content
 }
 for _, change := range patch.Files {
  if err := validatePath(change.Path); err != nil { return nil, err }
  current, exists := next[change.Path]
  switch change.Operation {
  case OperationCreate:
   if exists { return nil, fmt.Errorf("cannot create %q: file already exists", change.Path) }
   next[change.Path] = change.Content
  case OperationUpdate:
   if !exists { return nil, fmt.Errorf("cannot update %q: file does not exist", change.Path) }
   if ContentSHA(current) != change.ExpectedSHA { return nil, fmt.Errorf("cannot update %q: source hash mismatch", change.Path) }
   next[change.Path] = change.Content
  case OperationDelete:
   if !exists { return nil, fmt.Errorf("cannot delete %q: file does not exist", change.Path) }
   if ContentSHA(current) != change.ExpectedSHA { return nil, fmt.Errorf("cannot delete %q: source hash mismatch", change.Path) }
   delete(next, change.Path)
  default:
   return nil, fmt.Errorf("unsupported operation for %q", change.Path)
  }
 }
 return next, nil
}

func validatePath(name string) error {
 if strings.TrimSpace(name) == "" { return fmt.Errorf("file path is required") }
 if strings.ContainsRune(name, 0) || strings.Contains(name, "\\") { return fmt.Errorf("unsafe file path %q", name) }
 if strings.HasPrefix(name, "/") || path.IsAbs(name) { return fmt.Errorf("absolute file paths are not allowed: %q", name) }
 clean := path.Clean(name)
 if clean == "." || clean == ".." || strings.HasPrefix(clean, "../") { return fmt.Errorf("path escapes project root: %q", name) }
 if clean != name { return fmt.Errorf("file path must be normalized: %q", name) }
 return nil
}
