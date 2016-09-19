package match_test

import (
	. "github.com/iancmcc/jig/match"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var (
	paths = []string{
		"github.com/iancmcc/jig",
		"github.com/iancmcc/jig2",
		"github.com/iancmcc/jorg",
		"github.com/iandmcc/jig",
		"github.com/jort/jig",
		"gorthub.com/john/jig",
		"golang.x/blorgles/dinkum",
		"golang.x/blorgles/thing-dinkum",
	}
)

func testMatcher(query string) *SubstringPathMatcher {
	matcher := DefaultMatcher(query)
	for _, s := range paths {
		matcher.Add(s)
	}
	return matcher.(*SubstringPathMatcher)
}

var _ = Describe("DefaultPathMatcher", func() {

	It("should match an exact match", func() {
		results := testMatcher("github.com/iancmcc/jig").Match()
		Ω(results[0]).Should(Equal("github.com/iancmcc/jig"))
	})

	It("should match a name with a hyphen", func() {
		results := testMatcher("thing-dinkum").Match()
		Ω(results[0]).Should(Equal("golang.x/blorgles/thing-dinkum"))
	})

})
