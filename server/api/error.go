package api

import "net/http"

type Error struct {
	HttpStatusCode int      `json:"-"`
	Message        string   `json:"message"`
	Details        []string `json:"details,omitempty"`
}

func (e *Error) Error() string {
	return e.Message
}

func ErrBadRequest(message string, details []string) *Error {
	return &Error{HttpStatusCode: http.StatusBadRequest, Message: message, Details: details}
}

func ErrServiceUnavailable(message string) *Error {
	return &Error{HttpStatusCode: http.StatusServiceUnavailable, Message: message, Details: nil}
}

func ErrForbidden(message string) *Error {
	return &Error{HttpStatusCode: http.StatusForbidden, Message: message, Details: nil}
}

func ErrInternalServerError() *Error {
	return &Error{HttpStatusCode: http.StatusInternalServerError, Message: "Internal server error", Details: nil}
}
