package model

import (
	"database/sql"
	"time"
)

type Conversion struct {
	Id         int64
	Fullpath   string
	Path       string
	Filestem   string
	Ext        string
	ConvertTo  []ConvertTo
	IsDone     bool
	IsCanceled bool
	ErrorCode  int
	CreatedAt  time.Time
	UpdatedAt  sql.NullTime
}

type ConversionInfo struct {
	Fullpath       string
	Path           string
	Filestem       string
	Ext            string
	ConvertTo      []ConvertTo
	ReplaceOrigExt bool
}

type ConvertTo struct {
	Ext      string                 `json:"ext"`       // Required field
	ConvConf map[string]interface{} `json:"conv_conf"` // Optional conf with arbitrary fields
	Optional map[string]interface{} `json:"optional"`  // Catch-all for other fields
}

// UnmarshalYAML implements custom unmarshaling for ConvertTo
func (item *ConvertTo) UnmarshalYAML(unmarshal func(interface{}) error) error {
	// First, we unmarshal into a map to capture all fields
	rawData := make(map[string]interface{})
	if err := unmarshal(&rawData); err != nil {
		return err
	}

	// Now, attempt to extract the known fields
	if ext, ok := rawData["ext"]; ok {
		item.Ext = ext.(string) // Assuming ext is a string
		delete(rawData, "ext")
	}

	if conf, ok := rawData["conv_conf"]; ok {
		item.ConvConf = conf.(map[string]interface{}) // Assuming conf is a map
		delete(rawData, "conv_conf")
	}

	// Store the remaining fields (other than ext and conf) in the Optional map
	item.Optional = rawData

	return nil
}
