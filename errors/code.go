package errors

import (
	"fmt"
	"net/http"
	"sync"
)

var (
	unknownCoder ErrCode = ErrCode{1, http.StatusInternalServerError, "An internal server error occurred"} // 未注册的错误码，返回此信息
)

// Coder defines an interface for an error code detail information.
type Coder interface {
	// HTTP status that should be used for the associated error code.
	HTTPStatus() int

	// External (user) facing error text.
	String() string

	// Code returns the code of the coder
	Code() int
}

type ErrCode struct {
	// ErrCode refers to the integer code of the ErrCode.
	ErrCode int

	// HTTPCode status that should be used for the associated error code.
	HTTPCode int

	// External (user) facing error text.
	Msg string
}

func (coder ErrCode) Code() int {
	return coder.ErrCode

}

// String implements stringer. String returns the external error message,
// if any.
func (coder ErrCode) String() string {
	return coder.Msg
}

// HTTPStatus returns the associated HTTP status code, if any. Otherwise,
// returns 200.
func (coder ErrCode) HTTPStatus() int {
	if coder.HTTPCode == 0 {
		return 500
	}

	return coder.HTTPCode
}

// codes contains a map of error codes to metadata.
var codes = map[int]Coder{}
var codeMux = &sync.Mutex{}

// RegisterErrorCode registers a user-defined error code.
// Parameters:
//   - code: error code, must be a non-zero integer
//   - httpStatus: corresponding HTTP status code for API responses
//   - ext: external error message displayed to users
//
// Notes:
//   - If the error code already exists, it will override the existing definition
//   - Error code 0 is reserved for unknown errors and cannot be registered
//   - This function is thread-safe
//
// Example:
//
//	RegisterErrorCode(1001, http.StatusBadRequest, "Invalid parameters")
func RegisterErrorCode(code int, httpStatus int, ext string) {
	coder := ErrCode{code, httpStatus, ext}
	if coder.Code() == 0 {
		panic("code `0` is reserved as unknownCode error code")
	}

	codeMux.Lock()
	defer codeMux.Unlock()

	codes[coder.Code()] = coder
}

// MustRegisterErrorCode registers a user-defined error code.
// Parameters:
//   - code: error code, must be a non-zero integer
//   - httpStatus: corresponding HTTP status code for API responses
//   - ext: external error message displayed to users
//
// Notes:
//   - Panics if the error code already exists
//   - Error code 0 is reserved for unknown errors and cannot be registered
//   - This function is thread-safe
//   - Suitable for scenarios where error code uniqueness must be guaranteed
//
// Example:
//
//	MustRegisterErrorCode(1001, http.StatusBadRequest, "Invalid parameters")
func MustRegisterErrorCode(code int, httpStatus int, ext string) {
	coder := ErrCode{code, httpStatus, ext}
	if coder.Code() == 0 {
		panic("code '0' is reserved as ErrUnknown error code")
	}

	codeMux.Lock()
	defer codeMux.Unlock()

	if _, ok := codes[coder.Code()]; ok {
		panic(fmt.Sprintf("code: %d already exist", coder.Code()))
	}

	codes[coder.Code()] = coder
}

// Register register a user define error code.
// It will override the exist code.
func Register(coder Coder) {
	if coder.Code() == 0 {
		panic("code `0` is reserved as unknownCode error code")
	}

	codeMux.Lock()
	defer codeMux.Unlock()

	codes[coder.Code()] = coder
}

// MustRegister register a user define error code.
// It will panic when the same Code already exist.
func MustRegister(coder Coder) {
	if coder.Code() == 0 {
		panic("code '0' is reserved as ErrUnknown error code")
	}

	codeMux.Lock()
	defer codeMux.Unlock()

	if _, ok := codes[coder.Code()]; ok {
		panic(fmt.Sprintf("code: %d already exist", coder.Code()))
	}

	codes[coder.Code()] = coder
}

// ParseCoder parse any error into *withCode.
// nil error will return nil direct.
// None withStack error will be parsed as ErrUnknown.
func ParseCoder(err error) Coder {
	if err == nil {
		return nil
	}

	if v, ok := err.(*withCode); ok {
		if coder, ok := codes[v.code]; ok {
			return coder
		}
	}

	return unknownCoder
}

// IsCode reports whether any error in err's chain contains the given error code.
func IsCode(err error, code int) bool {
	if v, ok := err.(*withCode); ok {
		if v.code == code {
			return true
		}

		if v.err != nil {
			return IsCode(v.err, code)
		}

		return false
	}

	return false
}

func init() {
	codes[unknownCoder.Code()] = unknownCoder
}
