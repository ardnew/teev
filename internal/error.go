package internal

import (
	"fmt"
	"time"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

// MakeInvalidArgumentError returns a new Error with the given cause.
func MakeInvalidArgumentError(err ...error) Error {
	return MakeError(errors.New("invalid argument")).Wrap(err...)
}

// MakeInvalidOperationError returns a new Error with the given cause.
func MakeInvalidOperationError(err ...error) Error {
	return MakeError(errors.New("invalid operation")).Wrap(err...)
}

// ErrInvalidError is returned when the Error object itself is not valid.
var ErrInvalidError Error

// Error is the base type for module-specific errors.
//
// Every Error contains a single base error, called the "cause".
//
// One or more errors may be wrapped by an Error
// to express a chain or composition of errors,
// especially when multiple error conditions apply to a single operation.
//
// When initialized using a Make* constructor (or Wrap),
// Error records the context in which it was created
// including the callsite, stacktrace, and datetime.
//
// Many of the module's exported functions return an Error
// that wraps standard errors from the Go standard library.
//
// Error supports the [errors.Is] interface so that
// wrapped external errors can be compared directly.
//
// # Error Abstractions
//
// The following constructors return common errors that should be used as
// containers for more specific errors,
// or when the specific cause of an error is not immediately known:
//
//   - MakeInvalidArgumentError
//   - MakeInvalidOperationError
type Error struct {
	when   time.Time
	cause  error
	format Format
	wrap   []error
}

// MakeError returns a new Error with the given cause.
// The returned error contains the current datetime and a stacktrace
// relative to the time and location MakeError was called.
func MakeError(cause error) Error {
	return Error{when: time.Now(), cause: cause}
}

// MakeFormatError returns a new Error with the given cause and formatter.
// The returned error contains the current datetime and a stacktrace
// relative to the time and location MakeFormatError was called.
func MakeFormatError(cause error, format Format) Error {
	return Error{when: time.Now(), cause: cause, format: format}
}

// Unwrap returns the slice of all non-nil errors wrapped by e.
// If e contains no wrapped errors, Unwrap returns nil.
//
// In particular, Unwrap never returns an empty slice.
func (e Error) Unwrap() []error {
	if len(e.wrap) == 0 {
		return nil
	}
	return e.wrap
}

// Wrap replaces all wrapped errors in e
// with all non-nil errors in err.
// If err contains no non-nil errors,
// Wrap returns e with no errors wrapped.
func (e Error) Wrap(err ...error) Error {
	e.wrap = nil
	for _, x := range err {
		if x != nil {
			e.wrap = append(e.wrap, x)
		}
	}
	return e
}

// When returns the datetime when e was created.
func (e Error) When() time.Time {
	return e.when
}

// Cause returns the base error that caused e.
func (e Error) Cause() error {
	return e.cause
}

// Is reports whether the given error err is equivalent to e.
//
// The given error err is equivalent to e if either of the following are true:
//
//  1. e.Cause() is equal to err; or
//  2. err is an [Error], and e.Cause() is equal to err.Cause().
//
// In other words, Is compares the base errors — [Error.Cause] — of e and err,
// and it does not consider the wrapped error chains or the datetime of either.
//
// It must only compare the base errors (i.e., a "shallow" comparison)
// so that [errors.Is] can recursively shallow-compare all wrapped errors
// in a single pass.
//
// The Go doc comment on [errors.Is] states:
//
//	An Is method should only shallowly compare err and the target and not
//	call Unwrap on either.
func (e Error) Is(err error) bool {
	cmp := err
	if err, ok := err.(Error); ok {
		cmp = err.Cause()
	}
	return errors.Is(e.cause, cmp)
}

// Error returns a string representation of e.
func (e Error) Error() string {
	f := e.format
	if f == nil {
		f = FormatYAML
	}
	return f(e)
}

// See [errors.Frame.Format] for supported format strings.
func (e Error) formatStackTrace(frameFormat string) []string {
	type st interface{ StackTrace() errors.StackTrace }
	var frame []string
	if stack, ok := e.Cause().(st); ok {
		frame = make([]string, len(stack.StackTrace()))
		for i, x := range stack.StackTrace() {
			frame[i] = fmt.Sprintf(frameFormat, x)
		}
	}
	return frame
}

func (e Error) formatWrappedErrors() []string {
	err := make([]string, len(e.wrap))
	for i, x := range e.wrap {
		err[i] = x.Error()
	}
	return err
}

// Format functions return a formatted string representation of a given Error.
type Format func(Error) string

// FormatYAML returns a YAML-formatted string representation of err.
func FormatYAML(err Error) string {
	// We are going to use YAML to present the error data.
	// Hopefully this will alleviate all of the quoting and escaping
	// that you would get with nested JSON structures.
	enc, encErr := yaml.Marshal(
		struct {
			When  string   `yaml:"when"`
			What  string   `yaml:"what"`
			Where []string `yaml:"where,flow,omitempty"`
			Wrap  []string `yaml:"wrap,flow,omitempty"`
		}{
			When:  err.When().Format("2006-01-02 15:04:05"),
			What:  fmt.Sprintf("%v", err.Cause()),
			Where: err.formatStackTrace("%+v"),
			Wrap:  err.formatWrappedErrors(),
		},
	)
	if encErr != nil {
		panic(encErr)
	}
	return string(enc)
}

func UnformatYAML(err string) Error {
	var data struct {
		When  string   `yaml:"when"`
		What  string   `yaml:"what"`
		Where []string `yaml:"where,flow,omitempty"`
		Wrap  []string `yaml:"wrap,flow,omitempty"`
	}
	if err := yaml.Unmarshal([]byte(err), &data); err != nil {
		return MakeError(err)
	}
	return ErrInvalidError
}
