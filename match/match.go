package match

import (
	"sort"

	"github.com/iancmcc/jig/trie"
)

type Matcher interface {
	Add(s string)
	Match() []string
}

func DefaultMatcher(query string) Matcher {
	return &JaroWinklerPathMatcher{
		query:          query,
		minScore:       0.5,
		inbox:          make(chan string),
		boostThreshold: 0.7,
		prefixSize:     4,
		jwTrie:         trie.NewTrie(),
	}
}

type JaroWinklerScored struct {
	value string
	score float64
}

type ScoredArray []*JaroWinklerScored

func (s ScoredArray) Len() int           { return len(s) }
func (s ScoredArray) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s ScoredArray) Less(i, j int) bool { return s[i].score > s[j].score }
func (s ScoredArray) ToStringArray() (result []string) {
	for _, x := range s {
		result = append(result, x.value)
	}
	return
}

type JaroWinklerPathMatcher struct {
	query          string
	minScore       float64
	inbox          chan string
	boostThreshold float64
	prefixSize     int
	scores         ScoredArray
	jwTrie         *trie.CharSortedTrie
}

func (m *JaroWinklerPathMatcher) Add(s string) {
	//strings.Split(s, "/") // TODO: Use proper path separator
	m.jwTrie.Add(s)
	//score := smetrics.JaroWinkler(s, m.query, m.boostThreshold, m.prefixSize)
	//m.scores = append(m.scores, &JaroWinklerScored{s, score})
}

// Match returns the strings previously Added in sorted order
func (m *JaroWinklerPathMatcher) Match() []string {
	sort.Sort(m.scores)
	return m.scores.ToStringArray()
}
