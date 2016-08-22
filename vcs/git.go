package vcs

import "github.com/iancmcc/jig/manifest"

var (
	// Git is the singleton driver
	Git VCS = gitVCS{"git"}
)

// GitVCS is a git driver
type gitVCS struct {
	cmd string
}

func git(cmd ...string) <-chan Progress {
	return nil
}

// Clone satisfies the VCS interface
func (g *GitVCS) Clone(r manifest.Repo) <-chan Progress {
}

// Pull satisfies the VCS interface
func (g *GitVCS) Pull(r manifest.Repo) <-chan Progress {

}

// Checkout satisfies the VCS interface
func (g *GitVCS) Checkout(r manifest.Repo) <-chan Progress {

}
