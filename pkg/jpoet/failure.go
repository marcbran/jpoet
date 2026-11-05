package jpoet

import "github.com/marcbran/jpoet/internal/plugin"

type Failure = plugin.Failure

const FailureField = plugin.FailureField

func NewFailure(code int, message string) *Failure {
	return plugin.NewFailure(code, message)
}
