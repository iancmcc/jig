package fs_test

import (
	. "github.com/iancmcc/jig/fs"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

type testLister struct {
	files map[string][]string
}

func (l *testLister) ListChildren(path string) ([]string, error) {
	v, ok := l.files[path]
	if !ok {
		return []string{}, nil
	}
	return v, nil
}

var lister Lister = &testLister{
	map[string][]string{
		"/a":     {"b", "c"},
		"/a/b":   {"d", "e"},
		"/a/c":   {"d"},
		"/a/c/d": {"d"},
		"/a/b/d": {"e"},
	},
}

var _ = Describe("ParallelFinder", func() {

	var finder *ParallelFinder

	BeforeEach(func() {
		finder = &ParallelFinder{Lister: lister}
	})

	It("finds children of a given name", func() {
		var names []string
		for s := range finder.FindBelowNamed("/a", "d") {
			names = append(names, s)
		}
		Ω(names).Should(ConsistOf("/a/c/d", "/a/b/d", "/a/c/d/d"))
	})

	It("finds children with children of a given name", func() {
		var names []string
		for s := range finder.FindBelowWithChildrenNamed("/a", "d") {
			names = append(names, s)
		}
		Ω(names).Should(ConsistOf("/a/c", "/a/b", "/a/c/d"))
	})

})
