package trie

import (
	"fmt"
	"math"
)

type tuple struct {
	C []rune
	N *Node
	m int
}

func (t tuple) String() string {
	return fmt.Sprintf("<%s, %s, %d>", string(t.C), string(t.N.key), int(t.m))
}

func mNeeded(threshold, prefixLength, weight float64, ls, lt int) float64 {
	ls1, lt1 := float64(ls), float64(lt)
	lp := prefixLength * weight
	coefficient := threshold - lp - ((1.0 - lp) / 3.0)
	rightside := (3.0 * ls1 * lt1) / ((ls1 + lt1) * (1 - lp))
	return math.Ceil(coefficient * rightside)
}

func dMax(prefixLength, weight float64, ls, lt, m int) float64 {
	m1, ls1, lt1 := float64(m), float64(ls), float64(lt)
	lp := prefixLength * weight
	return ((1.0-lp)/3.0)*((m1/ls1)+(m1/lt1)+1) + lp
}

// Filter filters out candidates that will definitely have a Jaro-Winkler score
// below a given threshold by doing a depth-first traversal of the trie and
// checking matching characters
func (t *CharSortedTrie) Filter(s string, threshold float64) []string {
	var (
		item    tuple
		stack   []tuple
		matches []string
		u       = t.Key(s)
	)

	needed := mNeeded(threshold, t.prefixLength, t.weight, len(u)-1, t.minlen)

	for _, child := range t.root.children {
		stack = append(stack, tuple{u, child, 0})
	}
	for len(stack) > 0 {
		// Pop a tuple off the stack
		end := len(stack) - 1
		item, stack = stack[end], stack[:end]
		C, N, m := item.C, item.N, item.m
		mMax := math.Min(float64(len(C)), float64((t.maxlen+1)-N.level+1)) + float64(m)
		if mMax >= needed {
			if N.key <= C[0] {
				if N.key == C[0] {
					C = C[1:]
					m++
					if len(N.bucket) > 0 {
						dm := dMax(t.prefixLength, t.weight, N.level, (len(u) - 1), m)
						if dm >= threshold {
							matches = append(matches, N.bucket...)
						}
					}
				}
				for _, child := range N.children {
					stack = append(stack, tuple{C, child, m})
				}
			} else {
				C = C[1:]
				stack = append(stack, tuple{C, N, m})
			}
		}
	}
	return matches
}
