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
	FindBelowNamed(path, match string, depth int) <-chan string
	// FindBelowNamed finds all files below path that have children named match
	FindBelowWithChildrenNamed(path, match string, depth int) <-chan string
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
func (f *ParallelFinder) FindBelowNamed(path, match string, depth int) <-chan string {
	out := make(chan string)
	var walk func(d string, wdepth int)
	walk = func(d string, wdepth int) {
		defer f.wg.Done()
		names, err := f.Lister.ListChildren(d)
		if err != nil {
			return
		}
		for _, name := range names {
			p := filepath.Join(d, name)
			if match == name {
				out <- p
				if depth > 0 && wdepth >= depth {
					continue
				}
			}
			wdepth++
			f.wg.Add(1)
			go walk(p, wdepth)
		}
	}
	f.wg.Add(1)
	go walk(path, 1)
	go func() {
		defer close(out)
		f.wg.Wait()
	}()
	return out
}

// FindBelowWithChildrenNamed satisfies the Finder interface
func (f *ParallelFinder) FindBelowWithChildrenNamed(path, match string, depth int) <-chan string {
	out := make(chan string)
	var walk func(d string, wdepth int)
	walk = func(d string, wdepth int) {
		defer f.wg.Done()
		names, err := f.Lister.ListChildren(d)
		if err != nil {
			return
		}
		if contains(match, names) {
			out <- d
			if depth > 0 && wdepth >= depth {
				return
			}
			wdepth++

		}
		for _, name := range names {
			f.wg.Add(1)
			go walk(filepath.Join(d, name), wdepth)
		}
	}
	f.wg.Add(1)
	go walk(path, 1)
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
