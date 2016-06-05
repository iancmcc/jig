package trie_test

import (
	. "github.com/iancmcc/jig/trie"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("CharSortedTrie", func() {
	var t *CharSortedTrie

	BeforeEach(func() {
		t = NewTrie()
	})

	Context("a simple trie", func() {

		It("uses a key that is an ordered permutation of the value", func() {
			Ω(t.Key("hello")).Should(Equal([]rune("ehllo")))
		})

		It("allows retrieval of the original strings", func() {
			t.Add("since")
			t.Add("nices")
			t.Add("nice")
			Ω(t.Get("nices")).Should(ConsistOf("nices", "since"))
			Ω(t.Get("ecins")).Should(ConsistOf("nices", "since"))
			Ω(t.Get("nice")).Should(ConsistOf("nice"))
			Ω(t.Get("icen")).Should(ConsistOf("nice"))
			Ω(t.Get("icer")).Should(BeEmpty())
		})

	})

	Context("a Jaro-Winkler filter", func() {

		var (
			threshold    = float64(0.95)
			prefixLength = float64(4)
			weight       = float64(0.1)
		)

		It("should filter out below-threshold matches", func() {
			t.Add("mices")
			t.Add("nices")
			t.Add("niche")
			t.Add("niece")
			t.Add("since")

			result := t.Filter("nines", prefixLength, weight, threshold)

			Ω(result).Should(ConsistOf("nices", "since"))
		})
	})
})
