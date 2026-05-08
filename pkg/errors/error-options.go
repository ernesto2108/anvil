package errors

import "strings"

type (
	Option func(o *option)
	option struct {
		Errors
	}
)

// WithMessage set the errors message.
func WithMessage(msg ...string) Option {
	return func(o *option) {
		o.Message = strings.Join(msg, " ")
	}
}

// WithError set the errors message.
func WithError(err error) Option {
	return func(o *option) {
		o.ExternalError = err
	}
}

// WithErrorCode set the http error code.
func WithErrorCode(code int) Option {
	return func(o *option) {
		o.HttpErrorCode = code
	}
}
