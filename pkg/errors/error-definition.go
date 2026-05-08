package errors

import "net/http"

const (
	// General errors
	BadRequestErr              AnvilMeshCode = "BAD_REQUEST_ERR"
	InternalErr                AnvilMeshCode = "INTERNAL_ERR"
	NotFoundErr                AnvilMeshCode = "NOT_FOUND_ERR"
	UnauthorizedErr            AnvilMeshCode = "UNAUTHORIZED_ERR"
	ForbiddenErr               AnvilMeshCode = "FORBIDDEN_ERR"
	InsufficientPermissionsErr AnvilMeshCode = "INSUFFICIENT_PERMISSIONS_ERR"

	// Database errors
	QueryErr               AnvilMeshCode = "QUERY_ERR"
	ScanErr                AnvilMeshCode = "SCAN_ERR"
	TransactionErr         AnvilMeshCode = "TRANSACTION_ERR"
	NoChangesDetectedErr   AnvilMeshCode = "NO_CHANGES_ERR"
	TooManyRowsAffectedErr AnvilMeshCode = "TOO_MANYROWS_ERR"
	ForeignKeyViolation    AnvilMeshCode = "FOREIGN_KEY_VIOLATION"
	PermissionDenied       AnvilMeshCode = "PERMISSION_DENIED"
	UniqueViolation        AnvilMeshCode = "UNIQUE_VIOLATION"

	// JSON errors
	MarshalErr   AnvilMeshCode = "MARSHAL_ERR"
	UnmarshalErr AnvilMeshCode = "UNMARSHAL_ERR"

	// HTTP errors
	RequestHeadersErr AnvilMeshCode = "REQUEST_HEADERS_ERR"

	// Default error codes
	UnknownError        = "UNKNOWN_ERR"
	UnknownErrorMessage = "unknown error"
)

var definitions = map[AnvilMeshCode]Errors{
	BadRequestErr: {
		Code:          BadRequestErr,
		Message:       "bad request",
		HttpErrorCode: http.StatusBadRequest,
	},
	InternalErr: {
		Code:          InternalErr,
		Message:       "internal errors",
		HttpErrorCode: http.StatusInternalServerError,
	},
	NotFoundErr: {
		Code:          NotFoundErr,
		Message:       "not found",
		HttpErrorCode: http.StatusNotFound,
	},
	UnauthorizedErr: {
		Code:          UnauthorizedErr,
		Message:       "unauthorized",
		HttpErrorCode: http.StatusUnauthorized,
	},
	ForbiddenErr: {
		Code:          ForbiddenErr,
		Message:       "forbidden",
		HttpErrorCode: http.StatusForbidden,
	},
	InsufficientPermissionsErr: {
		Code:          InsufficientPermissionsErr,
		Message:       "insufficient permissions",
		HttpErrorCode: http.StatusForbidden,
	},
	QueryErr: {
		Code:          QueryErr,
		Message:       "query error",
		HttpErrorCode: http.StatusInternalServerError,
	},
	ScanErr: {
		Code:          ScanErr,
		Message:       "scan error",
		HttpErrorCode: http.StatusInternalServerError,
	},
	TransactionErr: {
		Code:          TransactionErr,
		Message:       "transaction error",
		HttpErrorCode: http.StatusInternalServerError,
	},
	NoChangesDetectedErr: {
		Code:          NoChangesDetectedErr,
		Message:       "no changes detected",
		HttpErrorCode: http.StatusInternalServerError,
	},
	TooManyRowsAffectedErr: {
		Code:          TooManyRowsAffectedErr,
		Message:       "too many rows affected",
		HttpErrorCode: http.StatusInternalServerError,
	},
	ForeignKeyViolation: {
		Code:          ForeignKeyViolation,
		Message:       "foreign key violation",
		HttpErrorCode: http.StatusInternalServerError,
	},
	PermissionDenied: {
		Code:          PermissionDenied,
		Message:       "permission denied",
		HttpErrorCode: http.StatusForbidden,
	},
	MarshalErr: {
		Code:          MarshalErr,
		Message:       "marshal error",
		HttpErrorCode: http.StatusInternalServerError,
	},
	UnmarshalErr: {
		Code:          UnmarshalErr,
		Message:       "unmarshal error",
		HttpErrorCode: http.StatusInternalServerError,
	},
	RequestHeadersErr: {
		Code:          RequestHeadersErr,
		Message:       "request headers error",
		HttpErrorCode: http.StatusBadRequest,
	},
	UnknownError: {
		Code:          UnknownError,
		Message:       UnknownErrorMessage,
		HttpErrorCode: http.StatusInternalServerError,
	},
}

func GetDefaultMessage(code AnvilMeshCode) Errors {
	if definition, ok := definitions[code]; ok {
		return definition
	}

	return definitions[UnknownError]
}
