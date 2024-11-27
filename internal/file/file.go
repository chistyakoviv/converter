package file

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func Exists(filePath string) bool {
	_, err := os.Stat(filePath)
	// Check if the error is because the file doesn't exist
	if os.IsNotExist(err) {
		return false
	}
	// If there's no error or a different kind of error, the file might exist
	return err == nil
}

func Trimwd(src string) (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return MakePathRelative(strings.TrimPrefix(src, wd)), nil
}

type FileInfo struct {
	Fullpath string
	Path     string
	Filestem string
	Ext      string
}

func ExtractInfo(src string) *FileInfo {
	// Extract the file name with extension
	fileName := filepath.Base(src)

	// Extract the file extension
	fileExt := filepath.Ext(src)

	return &FileInfo{
		Fullpath: src,
		Path:     filepath.Dir(src),
		Filestem: strings.TrimSuffix(fileName, fileExt),
		Ext:      strings.ToLower(strings.TrimPrefix(fileExt, ".")),
	}
}

func MakePathRelative(src string) string {
	if !strings.HasPrefix(src, "/") {
		return fmt.Sprintf("/%s", src)
	}
	return src
}
