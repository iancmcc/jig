package fs

import (
	"os"
	"path/filepath"
	"sync"
)

var empty []string

// Finder can find files or directories that match strings
type Finder interface {
	// FindBelowNamed finds all files below path that are named match
	FindBelowNamed(path, match string) <-chan string
	// FindBelowNamed finds all files below path that have children named match
	FindBelowWithChildrenNamed(path, match string) <-chan string
}

// Lister lists children below a given path
type Lister interface {
	// ListChildren lists children below a given path
	ListChildren(path string) ([]string, error)
}

// basicLister is an extremely simple directory lister
type basicLister struct{}

// ListChildren satisfies the Lister interface. The code is a port from Go
func (l *basicLister) ListChildren(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	stat, err := f.Stat()
	if err != nil {
		return nil, err
	}
	if !stat.IsDir() {
		return empty, err
	}
	names, err := f.Readdirnames(-1)
	f.Close()
	if err != nil {
		return nil, err
	}
	return names, nil
}

// ParallelFinder finds files in parallel
type ParallelFinder struct {
	Lister Lister
	wg     sync.WaitGroup
}

// FindBelowNamed satisfies the Finder interface
func (f *ParallelFinder) FindBelowNamed(path, match string) <-chan string {
	out := make(chan string)
	var walk func(d string)
	walk = func(d string) {
		defer f.wg.Done()
		names, err := f.Lister.ListChildren(d)
		if err != nil {
			return
		}
		for _, name := range names {
			f.wg.Add(1)
			p := filepath.Join(d, name)
			if match == name {
				out <- p
			}
			go walk(p)
		}
	}
	f.wg.Add(1)
	go walk(path)
	go func() {
		defer close(out)
		f.wg.Wait()
	}()
	return out
}

// FindBelowWithChildrenNamed satisfies the Finder interface
func (f *ParallelFinder) FindBelowWithChildrenNamed(path, match string) <-chan string {
	out := make(chan string)
	var walk func(d string)
	walk = func(d string) {
		defer f.wg.Done()
		names, err := f.Lister.ListChildren(d)
		if err != nil {
			return
		}
		if contains(match, names) {
			out <- d
		}
		for _, name := range names {
			f.wg.Add(1)
			go walk(filepath.Join(d, name))
		}
	}
	f.wg.Add(1)
	go walk(path)
	go func() {
		defer close(out)
		f.wg.Wait()
	}()
	return out
}

func contains(filter string, candidates []string) bool {
	for _, s := range candidates {
		if s == filter {
			return true
		}
	}
	return false
}
