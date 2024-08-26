package storage

import (
	"os"

	"github.com/pkg/errors"
)

type LocalStorage struct {
	file string
}

func NewLocalStorage(file string) *LocalStorage {
	return &LocalStorage{file: file}
}

func (s *LocalStorage) Save(mnemonic string) error {
	if fileExists(s.file) {
		return errors.New("file already exists")
	}
	f, err := os.Create(s.file)
	if err!= nil {
		return errors.Wrap(err, "failed to create file")
	}
	defer f.Close()

	f.WriteString(mnemonic)
	return nil
}

func(s *LocalStorage) Load() (string, error) {
	if !fileExists(s.file) {
		return "", errors.New("file not found")
	}

	data, err := os.ReadFile(s.file)
	if err!= nil {
		return "", errors.Wrap(err, "failed to read file")
	}
	return string(data), nil
}

func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return false
}