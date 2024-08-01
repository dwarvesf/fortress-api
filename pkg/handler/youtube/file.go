package youtube

import (
	"os"
)

func (h *handler) createFile(path, fileName, content string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := os.MkdirAll(path, os.ModePerm); err != nil {
			return err
		}
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
