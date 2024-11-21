package file

import (
	"fmt"
	"os"

	"github.com/h2non/filetype"
)

const headSize = 261

func readHead(src string) ([]byte, error) {
	wd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	file, err := os.Open(fmt.Sprintf("%s%s", wd, src))
	if err != nil {
		return nil, err
	}
	defer file.Close()

	head := make([]byte, headSize)
	file.Read(head)

	return head, nil
}

func IsImage(src string) (bool, error) {
	head, err := readHead(src)
	if err != nil {
		return false, err
	}

	return filetype.IsImage(head), nil
}

func IsVideo(src string) (bool, error) {
	head, err := readHead(src)
	if err != nil {
		return false, err
	}

	return filetype.IsVideo(head), nil
}
