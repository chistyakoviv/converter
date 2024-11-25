package deletionq

import "errors"

var (
	ErrPathAlreadyExist = errors.New("file with the specified path already exists")
)

const (
	ErrFailedToRemoveFile uint32 = iota + 1
)
