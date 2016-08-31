package vcs

import (
	"os"
	"path/filepath"

	"github.com/iancmcc/jig/config"
)

// VCS represents a version control system
type VCS interface {
	Clone(r *config.Repo, dir string) (<-chan Progress, error)
	Pull(r *config.Repo, dir string) (<-chan Progress, error)
	Checkout(r *config.Repo, dir string) (<-chan Progress, error)
}

func ApplyRepoConfig(root string, vcs VCS, repo *config.Repo) (<-chan Progress, <-chan Progress, error) {
	dir, err := RepoToPath(repo.Repo)
	if err != nil {
		return nil, nil, err
	}
	dir = filepath.Join(root, dir)
	dir, err = filepath.Abs(dir)
	if err != nil {
		return nil, nil, err
	}

	out := make(chan Progress)
	checkout := make(chan Progress)

	done := make(chan bool)

	go func(dir string) error {
		defer close(done)
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
		return nil
	}(dir)

	go func(dir string) error {
		defer close(checkout)
		<-done
		cochan, err := vcs.Checkout(repo, dir)
		if err != nil {
			return err
		}
		for p := range cochan {
			checkout <- p
		}
		return nil
	}(dir)

	return out, checkout, nil

}
