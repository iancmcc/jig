package trie

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

var (
	// ErrNoSuchNode is returned when there's no such node
	ErrNoSuchNode = errors.New("No such node")
)

// NewTrie returns a new Trie
func NewTrie() *CharSortedTrie {
	return &CharSortedTrie{
		root: &Node{
			children: make(map[rune]*Node),
		},
		prefixLength: 4,
		weight:       0.1,
	}
}

type runeSorter []rune

func (r runeSorter) Less(i, j int) bool { return r[i] < r[j] }
func (r runeSorter) Swap(i, j int)      { r[i], r[j] = r[j], r[i] }
func (r runeSorter) Len() int           { return len(r) }

// CharSortedTrie makes a trie
type CharSortedTrie struct {
	root         *Node
	maxlen       int
	minlen       int
	prefixLength float64
	weight       float64
}

func (t *CharSortedTrie) Print() {
	t.root.Print(0)
}

// Key returns the path of nodes under which this string would be stored
func (t *CharSortedTrie) Key(s string) []rune {
	r := runeSorter(s)
	sort.Sort(r)
	return r
}

// Add adds a string
func (t *CharSortedTrie) Add(s, orig string) {
	t.root.Add(t.Key(s), s, orig)
	l := len(s) - 1
	if l > t.maxlen {
		t.maxlen = l
	}
	if t.minlen == 0 || l < t.minlen {
		t.minlen = l
	}
}

// Get gets the strings associated with a key
func (t *CharSortedTrie) Get(s string) []Value {
	return t.root.Get(t.Key(s))
}

type Value struct {
	Key   string
	Value string
}

// Node is a node in a trie
type Node struct {
	key      rune
	children map[rune]*Node
	bucket   []Value
	level    int
}

func (n *Node) Print(level int) {
	fmt.Println(strings.Repeat(" ", level), string(n.key))
	for _, s := range n.bucket {
		fmt.Println(strings.Repeat(" ", level), "--", s)
	}
	for _, child := range n.children {
		child.Print(level + 1)
	}
}

// Add adds a value to the trie node
func (n *Node) Add(value []rune, segment, orig string) {
	if len(value) == 0 {
		n.bucket = append(n.bucket, Value{segment, orig})
		return
	}
	key, rest := value[0], value[1:]
	next, ok := n.children[key]
	if !ok {
		next = &Node{
			key:      key,
			children: make(map[rune]*Node),
			level:    n.level + 1,
		}
		n.children[key] = next
	}
	next.Add(rest, segment, orig)
}

func (n *Node) traverse(value []rune) (*Node, error) {
	if len(value) == 0 {
		return n, nil
	}
	next, ok := n.children[value[0]]
	if !ok {
		return nil, ErrNoSuchNode
	}
	return next.traverse(value[1:])
}

// Get gets the strings associated with a source string
func (n *Node) Get(value []rune) (result []Value) {
	if node, err := n.traverse(value); err == nil {
		result = append(result, node.bucket...)
	}
	return
}
