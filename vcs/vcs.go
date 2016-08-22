package vcs

import "github.com/iancmcc/jig/manifest"

// VCS represents a version control system
type VCS interface {
	Clone(r manifest.Repo) <-chan Progress
	Pull(r manifest.Repo) <-chan Progress
	Checkout(r manifest.Repo) <-chan Progress
}

// Progress is a unit of progress reported by VCS
type Progress struct {
	Label   string
	Current int
	Total   int
}
