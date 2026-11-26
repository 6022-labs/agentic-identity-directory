package errors

import (
	"net/http"
	"unicode/utf8"
)

// maxCauseLength caps a logged cause: a remote dependency can return an arbitrarily long error.
const maxCauseLength = 2048

// ClassifiedError is how every transport reports an application error; Cause is logged, never returned.
type ClassifiedError struct {
	Status int
	Code   string
	Body   any
	Cause  error
}

// NewClassifiedError maps an application error once for every transport; an unknown error becomes internal.
func NewClassifiedError(err error) *ClassifiedError {
	switch e := err.(type) {
	case *ValidationError:
		return &ClassifiedError{Status: http.StatusBadRequest, Body: e}
	case *BadRequestError:
		return &ClassifiedError{Status: http.StatusBadRequest, Code: e.Code, Body: e}
	case *NotFoundError:
		return &ClassifiedError{Status: http.StatusNotFound, Code: e.Code, Body: e}
	case *UnavailableError:
		return &ClassifiedError{Status: http.StatusBadGateway, Code: e.Code, Body: e, Cause: e.Cause}
	case *InternalError:
		return &ClassifiedError{Status: http.StatusInternalServerError, Code: e.Code, Body: e, Cause: e.Cause}
	default:
		return NewClassifiedError(NewInternalError("internal_error", err))
	}
}

// Failed reports a server-side failure rather than a rejected request.
func (c *ClassifiedError) Failed() bool {
	return c.Status >= http.StatusInternalServerError
}

// CauseMessage is the cause's text cut on a rune boundary, empty without a cause.
func (c *ClassifiedError) CauseMessage() string {
	if c.Cause == nil {
		return ""
	}
	message := c.Cause.Error()
	if len(message) <= maxCauseLength {
		return message
	}
	cut := maxCauseLength
	for cut > 0 && !utf8.RuneStart(message[cut]) {
		cut--
	}
	return message[:cut]
}
