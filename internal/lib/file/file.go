package file

import (
	"os"
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
