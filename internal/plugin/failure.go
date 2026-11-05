package plugin

import (
	"errors"
	"net/http"
)

const FailureField = "_failure"

type Failure struct {
	Code    int    `json:"code"`
	Reason  string `json:"reason"`
	Message string `json:"message"`
}

func NewFailure(code int, message string) *Failure {
	return &Failure{Code: code, Reason: http.StatusText(code), Message: message}
}

func (f *Failure) Error() string {
	return f.Message
}

func (f *Failure) Value() map[string]any {
	reason := f.Reason
	if reason == "" {
		reason = http.StatusText(f.Code)
	}
	return map[string]any{
		FailureField: map[string]any{
			"code":    float64(f.Code),
			"reason":  reason,
			"message": f.Message,
		},
	}
}

func asFailure(err error) (*Failure, bool) {
	var f *Failure
	if errors.As(err, &f) {
		return f, true
	}
	return nil, false
}
