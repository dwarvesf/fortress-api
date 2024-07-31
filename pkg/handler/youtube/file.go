package youtube

import (
	"os"
)

func (h *handler) createFile(path, fileName, content string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		os.MkdirAll(path, os.ModePerm)
	}

	newFile := path + "/" + fileName
	file, err := os.Create(newFile)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(content)
	return err
}
