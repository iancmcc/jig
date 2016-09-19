package match

import (
	"fmt"
	"sort"
	"strings"

	"github.com/xrash/smetrics"
)

type Matcher interface {
	Add(s string)
	Match() []string
}

func DefaultMatcher(query string) Matcher {
	query = strings.Map(func(r rune) rune {
		switch {
		case r == '/' || r == '-' || r == '.':
			return -1
		default:
			return r
		}
	}, query)
	return &SubstringPathMatcher{query: query}
	/*
		return &LevenshteinPathMatcher{
			query:            query,
			insertionCost:    1,
			deletionCost:     1,
			substitutionCost: 2,
		}
	*/
	/*
		return &JaroWinklerPathMatcher{
			query:          query,
			minScore:       0.5,
			inbox:          make(chan string),
			boostThreshold: 0.7,
			prefixSize:     4,
		}
	*/
}

type Scored struct {
	value string
	score float64
}

type ScoredArray []*Scored

func (s ScoredArray) Len() int           { return len(s) }
func (s ScoredArray) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s ScoredArray) Less(i, j int) bool { return s[i].score > s[j].score }
func (s ScoredArray) ToStringArray() (result []string) {
	for _, x := range s {
		result = append(result, x.value)
	}
	return
}

type Value struct {
	key   string
	value string
	level int
}

type LevenshteinPathMatcher struct {
	allstrings       []Value
	query            string
	insertionCost    int
	deletionCost     int
	substitutionCost int
}

type JaroWinklerPathMatcher struct {
	allstrings     []Value
	query          string
	minScore       float64
	inbox          chan string
	boostThreshold float64
	prefixSize     int
	scores         ScoredArray
}

type SubstringPathMatcher struct {
	allstrings []Value
	query      string
}

func Tokenize(s string) []string {
	return strings.FieldsFunc(s, func(r rune) bool {
		return r == '/' || r == '-' || r == '.'
	})
}

func (m *JaroWinklerPathMatcher) Add(s string) {
	segments := Tokenize(s)
	var (
		path   string
		maxlen = len(segments) - 1
	)
	for i := maxlen; i >= 0; i-- {
		if len(segments[i]) > 0 {
			path = segments[i] + path
			m.allstrings = append(m.allstrings, Value{s, path, maxlen - i})
		}
	}
}

// Match returns the strings previously Added in sorted order
func (m *JaroWinklerPathMatcher) Match() []string {
	results := map[string]*Scored{}
	for _, value := range m.allstrings {
		score := smetrics.JaroWinkler(m.query, value.value, m.boostThreshold, m.prefixSize)
		v, ok := results[value.key]
		if (!ok || v.score < score) && score >= m.minScore {
			results[value.key] = &Scored{value.key, score}
		}
	}
	for _, v := range results {
		m.scores = append(m.scores, v)
	}
	sort.Sort(m.scores)
	return m.scores.ToStringArray()
}

func (m *LevenshteinPathMatcher) Match() []string {
	results := map[string]*Scored{}
	for _, value := range m.allstrings {
		score := -float64(smetrics.WagnerFischer(m.query, value.value, m.insertionCost, m.deletionCost, m.substitutionCost))
		v, ok := results[value.key]
		if !ok || v.score < score {

			results[value.key] = &Scored{value.key, score}
		}
	}
	var scores ScoredArray
	for _, v := range results {
		scores = append(scores, v)
	}
	sort.Sort(scores)
	return scores.ToStringArray()
}

func (m *LevenshteinPathMatcher) Add(s string) {
	segments := Tokenize(s)
	var (
		path   string
		maxlen = len(segments) - 1
	)
	for i := maxlen; i >= 0; i-- {
		if len(segments[i]) > 0 {
			path = segments[i] + path
			m.allstrings = append(m.allstrings, Value{s, path, maxlen - i})
		}
	}
}

func (m *SubstringPathMatcher) Match() []string {
	results := map[string]*Scored{}
	for _, value := range m.allstrings {
		if !strings.Contains(value.value, m.query) {
			continue
		}
		if _, ok := results[value.key]; !ok {
			results[value.key] = &Scored{value.key, 1}
		}
	}
	var scores ScoredArray
	for _, v := range results {
		scores = append(scores, v)
	}
	sort.Sort(scores)
	return scores.ToStringArray()
}

func (m *SubstringPathMatcher) Add(s string) {
	segments := Tokenize(s)
	var (
		path   string
		maxlen = len(segments) - 1
	)
	for i := maxlen; i >= 0; i-- {
		if len(segments[i]) > 0 {
			path = segments[i] + path
			m.allstrings = append(m.allstrings, Value{s, path, maxlen - i})
		}
	}
	fmt.Println(m.allstrings)
}
