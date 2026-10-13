package errors

import "fmt"

type Code string

const (
 CodeInvalidArgument Code = "invalid_argument"
 CodeInvalidState Code = "invalid_state"
 CodeNotFound Code = "not_found"
)

type ValidationError struct { Code Code; Field string; Msg string }

func (e ValidationError) Error() string {
 if e.Field == "" { return fmt.Sprintf("%s: %s", e.Code, e.Msg) }
 return fmt.Sprintf("%s: %s: %s", e.Code, e.Field, e.Msg)
}

func Invalid(field, msg string) error { return ValidationError{Code: CodeInvalidArgument, Field: field, Msg: msg} }
