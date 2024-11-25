package model

import (
	"database/sql"
	"time"
)

type Deletion struct {
	Id         int64
	Fullpath   string
	IsDone     bool
	IsCanceled bool
	ErrorCode  int
	CreatedAt  time.Time
	UpdatedAt  sql.NullTime
}

type DeletionInfo struct {
	Fullpath string
}
