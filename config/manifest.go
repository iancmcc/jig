package config

import (
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

var (
	ManifestName  = "manifest.json"
	ErrNoManifest = errors.New("No manifest exists")
)

// Manifest represents a serialized description of the repositories to check
// out
type Manifest struct {
	Repos []*Repo
}

type Repo struct {
	Repo string
	Ref  string
}

// FromJSON creates a Manifest from a JSON reader
func FromJSON(r io.Reader) (*Manifest, error) {
	var m Manifest
	data, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(data, &m.Repos); err != nil {
		return nil, err
	}
	return &m, nil
}

func (m *Manifest) ToJSON(w io.Writer) error {
	enc := json.NewEncoder(w)
	return enc.Encode(m.Repos)
}

func (m *Manifest) Save(dir string) error {
	path, err := ManifestPath(dir)
	if err != nil {
		return err
	}

	tmp := path + "~"
	defer os.Remove(tmp)

	tmpfile, err := os.Create(tmp)
	if err != nil {
		return err
	}
	defer tmpfile.Close()

	if err := m.ToJSON(tmpfile); err != nil {
		return err
	}

	if err := tmpfile.Close(); err != nil {
		return err
	}

	return os.Rename(tmp, path)
}

func ManifestPath(dir string) (string, error) {
	root, err := FindClosestJigRoot(dir)
	if err != nil {
		return "", err
	}
	return filepath.Join(root, JigDirName, ManifestName), nil
}

func DefaultManifest(dir string) (*Manifest, error) {
	path, err := ManifestPath(dir)
	if err != nil {
		return nil, err
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return FromJSON(file)
}
