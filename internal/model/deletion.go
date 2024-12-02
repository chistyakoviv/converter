package model

import (
	"database/sql"
	"time"
)

const (
	DeletionStatusPending  = 0
	DeletionStatusDone     = 1
	DeletionStatusCanceled = 2
)

type Deletion struct {
	Id        int64
	Fullpath  string
	Status    int
	ErrorCode int
	CreatedAt time.Time
	UpdatedAt sql.NullTime
}

func (c *Deletion) IsDone() bool {
	return c.Status == DeletionStatusDone
}

func (c *Deletion) IsCanceled() bool {
	return c.Status == DeletionStatusCanceled
}

func (c *Deletion) IsPending() bool {
	return c.Status == DeletionStatusPending
}

type DeletionInfo struct {
	Fullpath string
}
