package vcs

import "github.com/iancmcc/jig/config"

// VCS represents a version control system
type VCS interface {
	Clone(r config.Repo) <-chan Progress
	Pull(r config.Repo) <-chan Progress
	Checkout(r config.Repo) <-chan Progress
}
