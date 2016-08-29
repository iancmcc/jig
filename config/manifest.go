package config

import (
	"encoding/json"
	"io"
	"io/ioutil"
)

// Manifest represents a serialized description of the repositories to check
// out
type Manifest struct {
	Repos []Repo
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
	return enc.Encode(m)
}
