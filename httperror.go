package aio

import (
	"net/http"
)

var (
	ErrBadRequest          error = newProblemDetails(http.StatusBadRequest)
	ErrInternalServerError error = newProblemDetails(http.StatusInternalServerError)
)
