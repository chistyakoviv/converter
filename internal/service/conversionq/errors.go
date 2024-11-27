package conversionq

import "errors"

var (
	ErrPathAlreadyExist     = errors.New("file with the specified path already exists")
	ErrFilestemAlreadyExist = errors.New("file with the specified filestem already exists")
)
