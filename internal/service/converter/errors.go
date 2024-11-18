package converter

import "errors"

type conversionError struct {
	msg  string
	code uint32
}

const (
	ErrFileDoesNotExist uint32 = iota + 1
	ErrConversion
	ErrInvalidConversionFormat
	ErrPopFailed
)

func NewConversionError(msg string, code uint32) *conversionError {
	return &conversionError{
		code: code,
		msg:  msg,
	}
}

func (e *conversionError) Error() string {
	return e.msg
}

func (e *conversionError) Code() uint32 {
	return e.code
}

func (e *conversionError) Is(target error) bool {
	var ce *conversionError
	return errors.As(target, &ce)
}

func GetConversionError(err error) *conversionError {
	var ce *conversionError
	if errors.As(err, &ce) {
		return ce
	}
	return nil
}
