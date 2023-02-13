package httperr

import (
	"errors"
	"net/http"
)

func StdStatus(statusCode int) Error {
	return NewStd(statusCode, statusCode, http.StatusText(statusCode))
}

func Internal(err error) Error {
	return New(http.StatusInternalServerError, err)
}

func BadRequest(reason string) Error {
	return New(http.StatusBadGateway, errors.New(reason))
}
