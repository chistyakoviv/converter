package file

import (
	"fmt"
	"os"

	"github.com/h2non/filetype"
)

const headSize = 261

func IsImage(src string) (bool, error) {
	wd, err := os.Getwd()
	if err != nil {
		return false, err
	}
	file, err := os.Open(fmt.Sprintf("%s%s", wd, src))
	if err != nil {
		return false, err
	}

	head := make([]byte, headSize)
	file.Read(head)

	return filetype.IsImage(head), nil
}

func IsVideo(src string) (bool, error) {
	wd, err := os.Getwd()
	if err != nil {
		return false, err
	}
	file, err := os.Open(fmt.Sprintf("%s%s", wd, src))
	if err != nil {
		return false, err
	}

	head := make([]byte, headSize)
	file.Read(head)

	return filetype.IsVideo(head), nil
}
