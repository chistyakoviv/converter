package conversionq

import "errors"

var (
	ErrPathAlreadyExist        = errors.New("file with the specified path already exists")
	ErrFileDoesNotExist        = errors.New("file does not exist")
	ErrFileTypeNotSupported    = errors.New("file type not supported")
	ErrFailedDetermineFileType = errors.New("failed to determine file type")
	ErrInvalidConversionFormat = errors.New("cannot convert to the specified format")
	ErrEmptyTargetFormatList   = errors.New("target format list is empty")
)
