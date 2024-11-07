package model

import "time"

type Conversion struct {
	Id             int64
	Fullpath       string
	Path           string
	Filestem       string
	Ext            string
	ConvertTo      []string
	IsDone         bool
	IsCanceled     bool
	ReplaceOrigExt bool
	ErrorCode      int
	CreatedAt      time.Time
	UpdatedAt      time.Time
}
