package aio

import "net/http"

type ProblemDetails struct {
	Title  string         `json:"title"`
	Status int            `json:"status"`
	Detail map[string]any `json:"detail"`
}

func (pd ProblemDetails) Error() string {
	return pd.Title
}

func newProblemDetails(status int) ProblemDetails {
	return ProblemDetails{
		Title:  http.StatusText(status),
		Status: status,
		Detail: nil,
	}
}
