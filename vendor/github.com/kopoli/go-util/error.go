package util

import (
	"fmt"
	"io"
	"os"

	"github.com/juju/errors"
)

type ErrorHandler struct {
	Out             io.Writer
	PrintStackTrace bool
}

// The default error handler
var E = ErrorHandler{
	Out:             os.Stderr,
	PrintStackTrace: true,
}

// New creates a new error variable
func (e *ErrorHandler) New(format string, a ...interface{}) error {
	ret := errors.NewErr(format, a...)
	ret.SetLocation(1)
	return &ret
}

// Annotate increases context information to the error
func (e *ErrorHandler) Annotate(err error, a ...interface{}) error {
	ret := errors.Annotate(err, fmt.Sprint(a...))
	if err, ok := ret.(*errors.Err); ok {
		err.SetLocation(1)
		ret = err
	}
	return ret
}

// Print writes the error message to predefined io.Writer
func (e *ErrorHandler) Print(err error, a ...interface{}) {
	fmt.Fprint(e.Out, "Error: ")
	if len(a) > 0 {
		fmt.Fprint(e.Out, a...)
		fmt.Fprint(e.Out, ": ")
	}
	fmt.Fprintf(e.Out, "%s\n", err)

	if e.PrintStackTrace {
		fmt.Fprintln(e.Out, "Error stack:")
		fmt.Fprintln(e.Out, errors.ErrorStack(err))
	}
}

func (e *ErrorHandler) Panic(err error, a ...interface{}) {
	e.Print(err, a...)
	panic("Irrecoverable error")
}

// ErrorList is a list of errors that can be printed in a single go
type ErrorList struct {
	Message string
	errors  []error
}

// NewErrorList returns an initialized ErrorList
func NewErrorList(message string) (ret *ErrorList) {
	return &ErrorList{
		Message: message,
	}
}

// Append adds a new error to the list
func (e *ErrorList) Append(err error) {
	e.errors = append(e.errors, err)
}

// Error returns the error string generated from the list
func (e *ErrorList) Error() string {
	if len(e.errors) == 0 {
		return ""
	}

	ret := "Error: " + e.Message + ": "

	for i := range e.errors {
		ret = ret + fmt.Sprintf("Error %d: %s; ", i+1, e.errors[i])
	}

	return ret
}

// IsEmpty returns true if the error list is empty
func (e *ErrorList) IsEmpty() bool {
	return len(e.errors) == 0
}
