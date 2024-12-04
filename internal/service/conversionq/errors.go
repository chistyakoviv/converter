package conversionq

import "errors"

var (
	ErrPathAlreadyExist        = errors.New("file with the specified path already exists")
	ErrFilestemAlreadyExist    = errors.New("file with the specified filestem already exists")
	ErrFileDoesNotExist        = errors.New("file does not exist")
	ErrFileTypeNotSupported    = errors.New("file type not supported")
	ErrFailedDetermineFileType = errors.New("failed to determine file type")
	ErrInvalidConversion       = errors.New("cannot convert to the specified format")
)
