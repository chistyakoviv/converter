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

const (
	MediaTypeImage = 1
	MediaTypeVideo = 2
)

type Deletion struct {
	Id        int64
	Fullpath  string
	Status    int
	MediaType int
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

func (c *Deletion) IsImage() bool {
	return c.MediaType == MediaTypeImage
}

func (c *Deletion) IsVideo() bool {
	return c.MediaType == MediaTypeVideo
}

type DeletionInfo struct {
	Fullpath  string
	MediaType int
}
