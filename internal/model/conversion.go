package model

import (
	"database/sql"
	"time"
)

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
	UpdatedAt      sql.NullTime
}

type ConversionInfo struct {
	Fullpath       string
	Path           string
	Filestem       string
	Ext            string
	ConvertTo      []string
	ReplaceOrigExt bool
}
