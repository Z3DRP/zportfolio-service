package routes

import (
	"fmt"
	"net/http"
)

type ErrDecodeRequestBody struct {
	Operation string
	Err       error
}

func (e *ErrDecodeRequestBody) Error() string {
	return fmt.Sprintf("could not json decode %v request body:: %v", e.Operation, e.Err)
}

func (e *ErrDecodeRequestBody) Unwrap() error {
	return e.Err
}

func NewDecodeRequestBodyErr(op string, e error) *ErrDecodeRequestBody {
	return &ErrDecodeRequestBody{
		Operation: op,
		Err:       e,
	}
}

type ErrRequestTimeout struct {
	Request *http.Request
}

func (e *ErrRequestTimeout) Error() string {
	return fmt.Sprintf("request: %s method: %s timed out", e.Request.URL, e.Request.Method)
}

func NewRequestTimeoutErr(r *http.Request) *ErrRequestTimeout {
	return &ErrRequestTimeout{
		Request: r,
	}
}
