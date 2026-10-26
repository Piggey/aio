package aio

import (
	"net/http"
)

var (
	ErrBadRequest          error = Error(http.StatusBadRequest, http.StatusText(http.StatusBadRequest), nil)
	ErrInternalServerError error = Error(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError), nil)
)
