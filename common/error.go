package common

import "fmt"

// CynoError can wrap a generic error with our own messages.
type CynoError struct {
	Err error
	Msg string
}

// Error returns the error from our CynoError (making it an `error` object).
func (e *CynoError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s", e.Msg, e.Err.Error())
	}
	return e.Msg
}

// Error wraps a generic error with our own message.
func Error(err error, msg string, args ...interface{}) *CynoError {
	m := msg
	if len(args) > 0 {
		m = fmt.Sprintf(msg, args...)
	}

	return &CynoError{
		Err: err,
		Msg: m,
	}
}

// ErrorMsg returns our message as an error.
func ErrorMsg(msg string, args ...interface{}) *CynoError {
	return Error(nil, msg, args...)
}
