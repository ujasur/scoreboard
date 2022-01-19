package main

import (
	"errors"
	"fmt"
	"net/http"
)

var (
	errAuthInvalid   = &authError{msg: "wrong credentials, try again"}
	errAuthMissing   = &authError{msg: "authorization required"}
	errAuthTokenType = &authError{msg: "invalid token type"}
	errUnauthorized  = errors.New("operation is not authorized")
	errVoteRejected  = errors.New("vote rejected")
	errSessionClosed = errors.New("session closed")
	errSessionOpen   = errors.New("session is already open")
)

type authError struct {
	msg string
}

func (e *authError) Error() string {
	return e.msg
}

// systemError is an error used by API handlers
type systemError struct {
	err error
	msg string
}

type errClientError struct {
	msg string
}

func (e *errClientError) Error() string {
	return e.msg
}

func (se *systemError) Error() string {
	return se.msg
}

func writeAPIError(w http.ResponseWriter, err error) {
	switch v := err.(type) {
	case *authError:
		http.Error(w, v.Error(), http.StatusUnauthorized)
		return
	case *errClientError:
		writeJSONError(w, http.StatusBadRequest, err)
		return
	}

	switch err {
	case errUnauthorized:
		http.Error(w, err.Error(), http.StatusForbidden)
	case errVoteRejected:
		fallthrough
	case errSessionOpen:
		fallthrough
	case errSessionClosed:
		writeJSONError(w, http.StatusBadRequest, err)
	default:
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func writeJSONError(w http.ResponseWriter, statusCode int, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	w.Write([]byte(fmt.Sprintf("{\"error\": \"%s\"}", err.Error())))
}

func newClientError(msg string) *errClientError {
	return &errClientError{msg: msg}
}

func newSystemError(msg string) *systemError {
	return &systemError{msg: msg}
}
