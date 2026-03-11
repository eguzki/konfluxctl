package kubearchive

import (
	"fmt"
	"net/http"
	"strings"
)

type Err struct {
	code int
	err  string
}

func (e Err) Error() string {
	return fmt.Sprintf("error calling kubearchive system - reason: %s - code: %d", e.err, e.code)
}

func (e Err) Code() int {
	return e.code
}

// codeForError returns the HTTP status for a particular error.
func codeForError(err error) int {
	switch t := err.(type) {
	case Err:
		return t.Code()
	}
	// Unknown
	return -1
}

// codeForError returns the HTTP status for a particular error.
func messageForError(err error) string {
	switch t := err.(type) {
	case Err:
		return t.err
	}
	// Unknown
	return err.Error()
}

// IsNotFound returns true if the specified error was created by NewNotFound.
func IsNotFound(err error) bool {
	return codeForError(err) == http.StatusNotFound
}

// IsBadRequest determines if err is an error which indicates that the request is invalid.
func IsBadRequest(err error) bool {
	return codeForError(err) == http.StatusBadRequest
}

// IsUnauthorized determines if err is an error which indicates that the request is unauthorized and
// requires authentication by the user.
func IsUnauthorized(err error) bool {
	return codeForError(err) == http.StatusUnauthorized
}

// IsForbidden determines if err is an error which indicates that the request is forbidden and cannot
// be completed as requested.
func IsMultipleResourcesFound(err error) bool {
	return codeForError(err) == http.StatusInternalServerError &&
		strings.Contains(messageForError(err), "more than one resource found")
}

// IsForbidden determines if err is an error which indicates that the request is forbidden and cannot
// be completed as requested.
func IsForbidden(err error) bool {
	return codeForError(err) == http.StatusForbidden
}
