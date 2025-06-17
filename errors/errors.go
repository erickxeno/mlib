package errors

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/bytedance/sonic"
)

const (
	Separator = ","
)

func New(message string) error {
	return &fundamental{
		msg: message,
	}
}

func Errorf(format string, args ...interface{}) error {
	return &fundamental{
		msg: fmt.Sprintf(format, args...),
	}
}

type fundamental struct {
	msg string
}

func (f *fundamental) Error() string { return f.msg }

func (f *fundamental) Format(s fmt.State, verb rune) {
	format(s, verb, f)
}

// ------------------------------ WrapM ------------------------------
func WrapM(err error, msg string) error {
	return WrapWithMsg(err, msg)
}

func WrapMF(err error, format string, args ...interface{}) error {
	return WrapWithMsgF(err, format, args...)
}

func WrapWithMsg(err error, msg string) error {
	if err == nil {
		return nil
	}
	return &withMessage{
		err: err,
		msg: msg,
	}
}

func WrapWithMsgF(err error, format string, args ...interface{}) error {
	if err == nil {
		return nil
	}
	return &withMessage{
		err: err,
		msg: fmt.Sprintf(format, args...),
	}
}

type withMessage struct {
	err error
	msg string
}

func (w *withMessage) Error() string { return fmt.Sprintf("%+v,%s", w.Cause(), w.msg) }
func (w *withMessage) Cause() error  { return w.err }

// Unwrap provides compatibility for Go 1.13 error chains.
func (w *withMessage) Unwrap() error { return w.err }

func (w *withMessage) Format(s fmt.State, verb rune) {
	format(s, verb, w)
}

// ------------------------------ WrapC ------------------------------

func WrapC(code int, err error) error {
	return WrapWithCode(code, err)
}

func WrapCF(code int, format string, args ...interface{}) error {
	return WrapWithCodeF(code, format, args...)
}

func WrapWithCode(code int, err error) error {
	if err == nil {
		return nil
	}
	return &withCode{
		code: code,
		err:  err,
	}
}

func WrapWithCodeF(code int, format string, args ...interface{}) error {
	return &withCode{
		code: code,
		err:  fmt.Errorf(format, args...),
	}
}

type withCode struct {
	code int
	err  error
}

// Error return the externally-safe error message.
func (w *withCode) Error() string { return fmt.Sprintf("%s", w) }

// Cause return the cause of the withCode error.
func (w *withCode) Cause() error { return w.err }

// Unwrap provides compatibility for Go 1.13 error chains.
func (w *withCode) Unwrap() error { return w.err }

func (w *withCode) Format(state fmt.State, verb rune) {
	format(state, verb, w, false)
}

// formatInfo contains all the error information.
type formatInfo struct {
	Code    *int    `json:"code,omitempty"`    // 错误码
	CodeDes *string `json:"message,omitempty"` // 错误码说明
	ErrDes  *string `json:"error,omitempty"`   // 错误信息
}

func format(state fmt.State, verb rune, w error, wrapPrintBefore ...bool) {
	switch verb {
	case 'v':
		str := bytes.NewBuffer([]byte{})
		jsonData := []*formatInfo{}

		var (
			flagDetail bool
			flagTrace  bool
			modeJSON   bool
		)

		if state.Flag('#') {
			modeJSON = true
			flagDetail = true
			flagTrace = true
		}

		if state.Flag('-') {
			flagDetail = true
		}
		if state.Flag('+') {
			flagTrace = true
		}

		sep := ""
		errs := list(w, false)
		if len(wrapPrintBefore) == 0 || !wrapPrintBefore[0] {
			// Reverse the slice
			for i, j := 0, len(errs)-1; i < j; i, j = i+1, j-1 {
				errs[i], errs[j] = errs[j], errs[i]
			}
		}

		for _, e := range errs {
			finfo := buildFormatInfo(e)
			jsonData, str = appendErrorFormat(jsonData, str, finfo, sep, flagDetail, flagTrace, modeJSON)
			sep = Separator

			if !flagTrace {
				break
			}

			if !flagDetail && !flagTrace && !modeJSON {
				break
			}
		}
		if modeJSON {
			var byts []byte
			byts, _ = sonic.Marshal(jsonData)

			_, _ = str.Write(byts)
		}

		fmt.Fprintf(state, "%s", strings.Trim(str.String(), "\r\n\t"))
	default:
		errs := list(w, false)
		if len(wrapPrintBefore) == 0 || !wrapPrintBefore[0] {
			// Reverse the slice
			for i, j := 0, len(errs)-1; i < j; i, j = i+1, j-1 {
				errs[i], errs[j] = errs[j], errs[i]
			}
		}

		sep := ""
		for _, e := range errs {
			finfo := buildFormatInfo(e)
			fmt.Fprintf(state, "%s%s", sep, finfo.String())
			sep = Separator
		}
	}
}

func appendErrorFormat(jsonData []*formatInfo, str *bytes.Buffer, finfo *formatInfo,
	sep string, flagDetail, flagTrace, modeJSON bool) ([]*formatInfo, *bytes.Buffer) {
	if modeJSON {
		jsonData = append(jsonData, finfo)
		return jsonData, str
	}

	// not modeJSON, print the error message to buffer
	if flagDetail || flagTrace {
		fmt.Fprintf(str, "%s%s", sep, finfo.String())
	} else {
		fmt.Fprintf(str, "%s%s", sep, finfo.ShortString()) // 只输出错误码和错误码说明，不输出内部详细的错误信息
	}
	return jsonData, str
}

// list will convert the error stack into a simple array.
func list(e error, skip bool) []error {
	ret := []error{}

	if e != nil {
		if w, ok := e.(interface{ Unwrap() error }); ok {
			if !skip {
				ret = append(ret, e)
			}
			_, iswithCode := e.(*withCode) // withCode 之所以跳过，是因为 withCode 的错误信息打印已经包含了里面一层的error信息
			ret = append(ret, list(w.Unwrap(), iswithCode)...)
		} else {
			if !skip {
				ret = append(ret, e)
			}
		}
	}

	return ret
}

func buildFormatInfo(e error) *formatInfo {
	var finfo *formatInfo

	switch err := e.(type) {
	case *fundamental:
		finfo = &formatInfo{
			ErrDes: &err.msg,
		}
	case *withCode:
		coder, ok := codes[err.code]
		if !ok {
			coder = unknownCoder
		}

		extMsg := coder.String()
		if extMsg == "" {
			extMsg = err.err.Error()
		}

		finfo = &formatInfo{
			Code:    toIntPtr(coder.Code()),
			CodeDes: toStringPtr(extMsg),
			ErrDes:  toStringPtr(err.err.Error()),
		}
	case *withMessage:
		finfo = &formatInfo{
			ErrDes: toStringPtr(err.msg),
		}
	default:
		finfo = &formatInfo{
			ErrDes: toStringPtr(err.Error()),
		}
	}

	return finfo
}

func (finfo *formatInfo) String() string {
	if finfo.Code == nil && finfo.CodeDes == nil {
		return *finfo.ErrDes
	}
	shortStr := finfo.ShortString()
	return fmt.Sprintf("%s%s%s", shortStr, Separator, *finfo.ErrDes)
}

func (finfo *formatInfo) ShortString() string {
	if finfo.Code == nil && finfo.CodeDes == nil {
		return *finfo.ErrDes // 返回内部错误信息
	}
	if finfo.Code == nil {
		return fmt.Sprintf("nilcode:%s", *finfo.CodeDes)
	} else if finfo.CodeDes == nil {
		return fmt.Sprintf("%d:nilcodeDes", *finfo.Code)
	}
	return fmt.Sprintf("%d:%s", *finfo.Code, *finfo.CodeDes)
}

// Cause returns the underlying cause of the error, if possible.
// An error value has a cause if it implements the following
// interface:
//
//	type causer interface {
//	       Cause() error
//	}
//
// If the error does not implement Cause, the original error will
// be returned. If the error is nil, nil will be returned without further
// investigation.
func Cause(err error) error {
	type causer interface {
		Cause() error
	}

	for err != nil {
		cause, ok := err.(causer)
		if !ok {
			break
		}

		if cause.Cause() == nil {
			break
		}

		err = cause.Cause()
	}
	return err
}

func toIntPtr(i int) *int {
	return &i
}

func toStringPtr(s string) *string {
	return &s
}
