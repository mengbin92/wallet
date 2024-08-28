package storage

import (
	"bufio"
	"fmt"
	"os"
	"strings"

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
	if err != nil {
		return errors.Wrap(err, "failed to create file")
	}
	defer f.Close()

	f.WriteString(mnemonic)
	return nil
}

func (s *LocalStorage) Load() (string, error) {
	if !fileExists(s.file) {
		return "", errors.New("file not found")
	}

	file, err := os.Open(s.file)
	if err != nil {
		return "", errors.Wrap(err, "failed to open file")
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	if scanner.Scan() {
		return scanner.Text(),nil
	}
	if err := scanner.Err(); err != nil {
		return "", errors.Wrap(err, "failed to scan file")
	}
	return "", errors.New("empty file")
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

func (s *LocalStorage) SaveKey(key string) error {
	file, err := os.OpenFile(s.file, os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return errors.Wrap(err, "failed to open file")
	}
	defer file.Close()

	_, err = file.WriteString(fmt.Sprintf("\nkey:%s", key))
	if err != nil {
		return errors.Wrap(err, "failed to write key to file")
	}
	return nil
}

func (s *LocalStorage) ListKeys() ([]string, error) {
	var lines []string

	file, err := os.Open(s.file)
	if err != nil {
		return nil, errors.Wrap(err, "failed to open file")
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "key:") {
			lines = append(lines, line[4:])
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, errors.Wrap(err, "failed to scan file")
	}
	return lines, nil
}
