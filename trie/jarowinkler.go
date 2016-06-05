package trie

import "math"

type stack []int

func (s stack) Empty() bool { return len(s) == 0 }
func (s stack) Peek() int   { return s[len(s)-1] }
func (s *stack) Put(i int)  { (*s) = append((*s), i) }
func (s *stack) Pop() int {
	d := (*s)[len(*s)-1]
	(*s) = (*s)[:len(*s)-1]
	return d
}

type tuple struct {
	C []rune
	N *Node
	m int
}

func mNeeded(threshold, prefixLength, weight float64, ls, lt int) float64 {
	coefficient := threshold - (prefixLength * weight) - ((1 - prefixLength*weight) / 3)
	rightside := float64(3*ls*lt) / (float64(ls+lt) * (1 - (prefixLength * weight)))
	return math.Ceil(coefficient * rightside)
}

func dMax(prefixLength, weight float64, ls, lt, m int) float64 {
	return ((1-(prefixLength*weight))/3.0)*float64((m/ls)+(m/lt)+1) + prefixLength*weight
}

// Filter filters out candidates that will be below a Jaro-Winkler threshold score.
func (t *CharSortedTrie) Filter(s string, prefixLength, weight, threshold float64) []string {
	var (
		stack   []tuple
		matches []string
		u       = t.Key(s)
	)

	stack = append(stack, tuple{u, t.root, 0})
	for len(stack) > 0 {
		// Pop a tuple off the stack
		end := len(stack) - 1
		s, stack := stack[end], stack[:end]
		C, N, m := s.C, s.N, s.m
		mMax := math.Min(float64(len(C)), float64(t.maxlen-N.level)) + float64(m)
		if mMax >= mNeeded(threshold, prefixLength, weight, len(C), t.minlen) {
			if N.key <= C[0] {
				if N.key == C[0] {
					C = C[1:]
					m++
					if len(N.bucket) > 0 && dMax(prefixLength, weight, N.level, len(u), m) >= threshold {
						matches = append(matches, N.bucket...)
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
