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
	}
)

func testMatcher(query string) *JaroWinklerPathMatcher {
	matcher := DefaultMatcher(query)
	for _, s := range paths {
		matcher.Add(s)
	}
	return matcher.(*JaroWinklerPathMatcher)
}

var _ = Describe("JaroWinklerPathMatcher", func() {

	It("should match an exact match", func() {
		results := testMatcher("github.com/iancmcc/jig").Match()
		Ω(results[0]).Should(Equal("github.com/iancmcc/jig"))
	})

	//It("should match by repository name", func() {
	//	results := testMatcher("jig").Match()
	//	Ω(results[:4]).Should(ConsistOf("github.com/iancmcc/jig", "github.com/iandmcc/jig", "github.com/jort/jig", "gorthub.com/john/jig"))
	//	Ω(results[5]).Should(Equal("github.com/iancmcc/jig2"))
	//})

})
