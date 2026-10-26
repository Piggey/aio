package middleware

import (
	"errors"
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/Piggey/aio"
)

// PanicError is returned by Recover when the rest of the chain panicked.
type PanicError struct {
	Value any
	Stack []byte
}

func (pe PanicError) Error() string {
	return fmt.Sprintf("panic: %v", pe.Value)
}

// Recover turns a panic further down the chain into a PanicError,
// which App.ErrorHandler responds to with 500 Internal Server Error.
// http.ErrAbortHandler is re-panicked, so net/http can abort the response.
func Recover() aio.Middleware {
	return func(next aio.HandlerFunc) aio.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) (err error) {
			defer func() {
				rec := recover()
				if rec == nil {
					return
				}

				if isAbortHandler(rec) {
					panic(rec)
				}

				err = PanicError{
					Value: rec,
					Stack: debug.Stack(),
				}
			}()

			return next(w, r)
		}
	}
}

func isAbortHandler(rec any) bool {
	recErr, ok := rec.(error)
	return ok && errors.Is(recErr, http.ErrAbortHandler)
}
