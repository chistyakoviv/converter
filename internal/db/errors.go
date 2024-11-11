package db

import "errors"

// error var should have name of the form ErrFoo (ST1012)go-staticcheck
var (
	ErrNotFound    = errors.New("not found")
	ErrEntryExists = errors.New("entry already exists")
)
