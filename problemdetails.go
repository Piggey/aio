package aio

import "fmt"

type ProblemDetails struct {
	Title  string `json:"title" xml:"title"`
	Status int    `json:"status" xml:"status"`
	Detail string `json:"detail,omitempty" xml:"detail,omitempty"`
}

func (pd ProblemDetails) Error() string {
	if pd.Detail == "" {
		return pd.Title
	}

	return fmt.Sprintf("%s: %s", pd.Title, pd.Detail)
}

func Error(status int, title string, err error) ProblemDetails {
	detail := ""
	if err != nil {
		detail = err.Error()
	}

	return ProblemDetails{
		Title:  title,
		Status: status,
		Detail: detail,
	}
}
