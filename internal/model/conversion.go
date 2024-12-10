package model

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/chistyakoviv/converter/internal/file"
)

const (
	ConversionStatusPending  = 0
	ConversionStatusDone     = 1
	ConversionStatusCanceled = 2
)

type Conversion struct {
	Id        int64
	Fullpath  string
	Path      string
	Filestem  string
	Ext       string
	ConvertTo []ConvertTo
	Status    int
	ErrorCode int
	CreatedAt time.Time
	UpdatedAt sql.NullTime
}

func (c *Conversion) IsDone() bool {
	return c.Status == ConversionStatusDone
}

func (c *Conversion) IsCanceled() bool {
	return c.Status == ConversionStatusCanceled
}

func (c *Conversion) IsPending() bool {
	return c.Status == ConversionStatusPending
}

// Since Go does not support optional parameters, a variadic parameter is used instead.
// If optionalPathPrefix is not provided or empty, the default path prefix will be the working directory.
func (c *Conversion) AbsoluteSourcePath(optionalPathPrefix ...string) (string, error) {
	pathPrefix, err := constructPathPrefix(optionalPathPrefix...)
	if err != nil {
		return "", err
	}
	return pathPrefix + c.Fullpath, nil
}

// Since Go does not support optional parameters, a variadic parameter is used instead.
// If optionalPathPrefix is not provided or empty, the default path prefix will be the working directory.
func (c *Conversion) AbsoluteDestinationPath(entry ConvertTo, optionalPathPrefix ...string) (string, error) {
	pathPrefix, err := constructPathPrefix(optionalPathPrefix...)
	if err != nil {
		return "", err
	}

	dest := fmt.Sprintf("%s%s/%s", pathPrefix, c.Path, c.Filestem)

	// Preserve the original extension if specified
	var hasReplaceOrigExt, isReplaceOrigExtBool, replaceOrigExt bool
	var replaceOrigExtValue interface{}
	if replaceOrigExtValue, hasReplaceOrigExt = entry.Optional["replace_orig_ext"]; hasReplaceOrigExt {
		replaceOrigExt, isReplaceOrigExtBool = replaceOrigExtValue.(bool)
	}
	if !hasReplaceOrigExt || !isReplaceOrigExtBool || !replaceOrigExt {
		// Append the original extension
		dest = dest + "." + c.Ext
	}

	// Add suffix if specified
	var hasSuffix, isSuffixStr bool
	var suffix string
	var suffixValue interface{}
	if suffixValue, hasSuffix = entry.Optional["suffix"]; hasSuffix {
		suffix, isSuffixStr = suffixValue.(string)
	}
	if hasSuffix && isSuffixStr {
		dest = dest + suffix
	}

	return dest + "." + entry.Ext, nil
}

type ConversionInfo struct {
	Fullpath  string
	Path      string
	Filestem  string
	Ext       string
	ConvertTo []ConvertTo
}

// There is no way to makke optional parameters, so use variadic parameter
func (c *ConversionInfo) AbsoluteSourcePath(optionalPathPrefix ...string) (string, error) {
	pathPrefix, err := constructPathPrefix(optionalPathPrefix...)
	if err != nil {
		return "", err
	}
	return pathPrefix + c.Fullpath, nil
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
	}

	if conf, ok := rawData["conv_conf"]; ok {
		item.ConvConf = conf.(map[string]interface{}) // Assuming conf is a map
	}

	if optional, ok := rawData["optional"]; ok {
		item.Optional = optional.(map[string]interface{}) // Assuming optional is a map
	}

	return nil
}

func (item *ConvertTo) UnmarshalENV(unmarshal func(interface{}) error) error {
	// First, we unmarshal into a map to capture all fields
	rawData := make(map[string]interface{})
	log.Fatalf("rawData: %v\n", rawData)
	if err := unmarshal(&rawData); err != nil {
		return err
	}

	// Now, attempt to extract the known fields
	if ext, ok := rawData["ext"]; ok {
		item.Ext = ext.(string) // Assuming ext is a string
	}

	if conf, ok := rawData["conv_conf"]; ok {
		item.ConvConf = conf.(map[string]interface{}) // Assuming conf is a map
	}

	if optional, ok := rawData["optional"]; ok {
		item.Optional = optional.(map[string]interface{}) // Assuming optional is a map
	}

	return nil
}

func constructPathPrefix(optionalPathPrefix ...string) (string, error) {
	var pathPrefix string
	if len(optionalPathPrefix) > 0 && optionalPathPrefix[0] != "" {
		pathPrefix = optionalPathPrefix[0]
	} else {
		wd, err := os.Getwd()
		if err != nil {
			return "", fmt.Errorf("failed to get working directory: %w", err)
		}
		pathPrefix = wd
	}
	return pathPrefix, nil
}

func ToConversionInfoFromFileInfo(finfo *file.FileInfo) *ConversionInfo {
	return &ConversionInfo{
		Fullpath: finfo.Fullpath,
		Path:     finfo.Path,
		Filestem: finfo.Filestem,
		Ext:      finfo.Ext,
	}
}
