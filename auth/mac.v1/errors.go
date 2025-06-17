package mac

import "github.com/erickxeno/mlib/errors"

var (
	ErrMissSuInfo      = errors.New("admin auth must specify a suInfo parameter")
	ErrMissSK          = errors.New("auth must specify a sk parameter")
	ErrMissAuthType    = errors.New("must specify a auth type parameter")
	ErrUnknownAuthType = errors.New("unknown auth type parameter")
	ErrMissAkSk        = errors.New("auth must specify a ak/sk parameter")
)
