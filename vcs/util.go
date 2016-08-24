package vcs

import (
	"fmt"
	"sync"
)

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

			// If this is the beginning of a new stage for this ob, clean up others
			if progress.IsBegin {
				for k := range seen {
					if k == msg {
						continue
					}
					if v, ok := states[k]; ok {
						delete(v, repo)
						fmt.Println(k, v)
						if len(v) == 0 {
							delete(states, k)
						}
					}
				}
			}

			// Set this value
			statemap[repo] = progress

			// Calculate the new total for the lowest one
			for _, loweststate := range order {
				if smap, ok := states[loweststate]; !ok {
					continue
				} else {
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

			}

		}
		close(resultchan)
	}()

	return resultchan
}
