package task

import "errors"

var (
	ErrScanAlreadyRunning = errors.New("scanning in progress")
)
