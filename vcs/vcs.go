package vcs

import (
	"context"

	"github.com/iancmcc/jig/manifest"
)

// VCS represents a version control system
type VCS interface {
	Clone(ctx context.Context, r manifest.Repo) <-chan Progress
	Pull(ctx context.Context, r manifest.Repo) <-chan Progress
	Checkout(ctx context.Context, r manifest.Repo) <-chan Progress
}

// Progress is a unit of progress reported by VCS
type Progress struct {
	IsBegin bool
	Message string
	Current int
	Total   int
}
