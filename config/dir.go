package config

import (
	"errors"
	"os"
	"path/filepath"
)

var (
	// ErrNoJigRoot is returned when no Jig root can be found
	ErrNoJigRoot = errors.New("Could not find Jig root")
)

const (
	JigDirName = ".jig"
)

func JigRootDir(path string) string {
	return filepath.Join(path, JigDirName)
}

func dirExists(path string) bool {
	if _, err := os.Stat(path); err != nil {
		return false
	}
	return true
}

func IsJigRoot(path string) bool {
	var err error
	if path, err = filepath.Abs(path); err != nil {
		return false
	}
	return dirExists(filepath.Join(path, JigDirName))
}

func FindClosestJigRoot(path string) (string, error) {
	var err error
	if path, err = filepath.Abs(path); err != nil {
		return "", ErrNoJigRoot
	}
	parent := path
	for parent != "/" {
		if IsJigRoot(parent) {
			return parent, nil
		}
		parent = filepath.Dir(parent)
	}
	return "", ErrNoJigRoot
}

func CreateJigRoot(path string) error {
	jd := filepath.Join(path, JigDirName)
	if !dirExists(jd) {
		return os.MkdirAll(jd, os.ModeDir|0755)
	}
	return nil
}
