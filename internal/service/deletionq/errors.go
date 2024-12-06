package deletionq

import "errors"

var (
	ErrPathAlreadyExist = errors.New("file with the specified path already exists")
	ErrFileDoesNotExist = errors.New("file does not exist")
)

const (
	ErrFailedToRemoveFile uint32 = iota + 1
)
