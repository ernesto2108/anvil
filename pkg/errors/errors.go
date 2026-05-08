package errors

import (
	"fmt"
	"log"
	"net/http"

	"google.golang.org/grpc/codes"
)

type (
	AnvilMeshCode string

	Errors struct {
		Code          AnvilMeshCode
		Message       string
		HttpErrorCode int
		GrpcErrorCode codes.Code
		ExternalError error
	}

	ErrorResponse struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	}
)

// New creates a new error with the given code and options.
func New(code AnvilMeshCode, opts ...Option) error {
	var (
		definition = GetDefaultMessage(code)
		opt        = option{}
	)

	for _, v := range opts {
		v(&opt)
	}

	definition.Code = code

	WithOptions(opt, &definition)

	return &definition
}

// WithOptions applies the given options to the error definition.
func WithOptions(opt option, definition *Errors) {
	if opt.Message != "" {
		definition.Message = opt.Message
	}

	if opt.HttpErrorCode > 0 {
		definition.HttpErrorCode = opt.HttpErrorCode
	}

	if opt.GrpcErrorCode > 0 {
		definition.GrpcErrorCode = opt.GrpcErrorCode
	}

	if opt.ExternalError != nil {
		definition.ExternalError = opt.ExternalError
	}
}

func (e *Errors) Error() string {
	if e.ExternalError != nil {
		return fmt.Sprintf("%s: %s", e.ExternalError, e.Code)
	}

	return e.Message
}

func getError(err error) *Errors {
	if err == nil {
		return nil
	}

	customErr, ok := err.(*Errors)
	if ok {
		return customErr
	}

	log.Printf("warn: error is not of type *Errors: %v", err)

	return &Errors{
		Code:    UnknownError,
		Message: err.Error(),
	}
}

// Code returns the AnvilMeshCode of the error.
func Code(err error) AnvilMeshCode {
	if err == nil {
		return ""
	}

	if customErr := getError(err); customErr != nil {
		return customErr.Code
	}

	return ""
}

// Message returns the message of the error.
func Message(err error) string {
	if err == nil {
		return ""
	}

	if customErr := getError(err); customErr != nil {
		return customErr.Message
	}

	return err.Error()
}

// HttpCode returns the HTTP status code of the error.
func HttpCode(err error) int {
	if err == nil {
		return http.StatusOK // Default to OK for nil error
	}

	if customErr := getError(err); customErr != nil {
		return customErr.HttpErrorCode
	}

	return http.StatusInternalServerError // Default for non-SaludMesh errors
}

// GrpcCode returns the gRPC status code of the error.
func GrpcCode(err error) codes.Code {
	if err == nil {
		return codes.OK
	}

	if customErr := getError(err); customErr != nil {
		return customErr.GrpcErrorCode
	}

	return codes.Internal // Default for non-SaludMesh errors
}

// ExternalError returns the underlying external error, if any.
func ExternalError(err error) error {
	if err == nil {
		return nil
	}

	if customErr := getError(err); customErr != nil {
		return customErr.ExternalError
	}

	return err
}

// ErrorLog logs the error if it is not nil.
func ErrorLog(err error) {
	if err != nil {
		log.Printf("ERROR: %v", err)
	}
}
