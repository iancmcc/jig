package vcs

import (
	"os"
	"path/filepath"

	"github.com/iancmcc/jig/config"
	"github.com/iancmcc/jig/utils"
)

// VCS represents a version control system
type VCS interface {
	Clone(r *config.Repo, dir string) (<-chan Progress, error)
	Pull(r *config.Repo, dir string) (<-chan Progress, error)
	Checkout(r *config.Repo, dir string) error
	Status(r *config.Repo, dir string) (*Status, error)
}

// Status is a function
type Status struct {
	Repo                        string
	OrigRef                     string
	Staged, Unstaged, Untracked bool
	Branch                      string
}

// ApplyRepoConfig is a function
func ApplyRepoConfig(root string, vcs VCS, repo *config.Repo) (<-chan Progress, error) {
	dir, err := utils.RepoToPath(repo.Repo)
	if err != nil {
		return nil, err
	}
	dir = filepath.Join(root, dir)
	dir, err = filepath.Abs(dir)
	if err != nil {
		return nil, err
	}

	out := make(chan Progress)

	go func(dir string) error {
		defer close(out)
		if _, err := os.Stat(dir); err != nil {
			// Directory doesn't exist
			clonechan, err := vcs.Clone(repo, dir)
			if err != nil {
				return err
			}
			for p := range clonechan {
				out <- p
			}
		} else {
			pullchan, err := vcs.Pull(repo, dir)
			if err != nil {
				return err
			}
			for p := range pullchan {
				out <- p
			}
		}
		vcs.Checkout(repo, dir)
		return nil
	}(dir)

	return out, nil
}
