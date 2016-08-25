package vcs

import (
	"fmt"
	"sync"
)

// Progress is a unit of progress reported by VCS
type Progress struct {
	Repo    string
	IsBegin bool
	IsEnd   bool
	Message string
	Current int
	Total   int
}

// CombinedProgress combines the progress from multiple operations into
// a single stream that reports on overall progress
func CombinedProgress(progs ...<-chan Progress) <-chan Progress {

	aggregate := make(chan Progress)
	resultchan := make(chan Progress)

	var wg sync.WaitGroup
	for _, ch := range progs {
		wg.Add(1)
		go func(c <-chan Progress) {
			defer wg.Done()
			for prog := range c {
				aggregate <- prog
			}
		}(ch)
	}
	go func() {
		wg.Wait()
		close(aggregate)
	}()

	go func() {

		seen := map[string]struct{}{}
		order := []string{}
		states := map[string]map[string]Progress{}

		for progress := range aggregate {
			var (
				statemap map[string]Progress
				ok       bool
			)
			msg := progress.Message
			repo := progress.Repo

			seen[msg] = struct{}{}

			// Get the map of current progress for this state
			if statemap, ok = states[msg]; !ok {
				statemap = map[string]Progress{}
				states[msg] = statemap
				order = append(order, msg)
			}

			// Set this value
			statemap[repo] = progress

			// Calculate the new total for the lowest one
			for _, loweststate := range order {
				smap, ok := states[loweststate]
				if !ok {
					continue
				}
				result := Progress{
					Message: loweststate,
					Repo:    fmt.Sprintf("%d repos", len(smap)),
				}
				for _, p := range smap {
					result.Total += p.Total
					result.Current += p.Current
				}
				resultchan <- result
				break

			}

			// If this is the end of a stage for this ob, clean up
			if progress.IsEnd {
				delete(statemap, repo)
				if len(statemap) == 0 {
					delete(states, msg)
				}
			}

		}
		close(resultchan)
	}()

	return resultchan
}
