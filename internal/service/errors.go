package service

import "errors"

type converterError struct {
	msg  string
	code uint32
}

// Conversion erros: 1 - 99
const (
	ErrFileDoesNotExist uint32 = iota + 1
	ErrUnableToConvertFile
	ErrInvalidConversionFormat
	ErrWrongSourceFile
)

// Deletion Errors: 100 - 199
const (
	ErrFailedToRemoveFile uint32 = iota + 100
	ErrFileQueuedForDeletion
)

func NewConverterError(msg string, code uint32) *converterError {
	return &converterError{
		code: code,
		msg:  msg,
	}
}

func (e *converterError) Error() string {
	return e.msg
}

func (e *converterError) Code() uint32 {
	return e.code
}

func (e *converterError) Is(target error) bool {
	var ce *converterError
	return errors.As(target, &ce)
}

func GetConverterError(err error) *converterError {
	var ce *converterError
	if errors.As(err, &ce) {
		return ce
	}
	return nil
}
