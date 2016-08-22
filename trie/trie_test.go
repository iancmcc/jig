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
			t.Add("since", "since")
			t.Add("nices", "nices")
			t.Add("nice", "nice")
			Ω(t.GetString("nices")).Should(ConsistOf("nices", "since"))
			Ω(t.GetString("ecins")).Should(ConsistOf("nices", "since"))
			Ω(t.GetString("nice")).Should(ConsistOf("nice"))
			Ω(t.GetString("icen")).Should(ConsistOf("nice"))
			Ω(t.GetString("icer")).Should(BeEmpty())
		})

	})

	Context("a Jaro-Winkler filter", func() {

		var (
			threshold = float64(0.9)
		)

		It("should filter out below-threshold matches", func() {
			t.Add("mices", "mices")
			t.Add("nices", "nices")
			t.Add("niche", "niche")
			t.Add("niece", "niece")
			t.Add("since", "since")

			result := t.FilterString("nines", threshold)

			Ω(result).Should(ConsistOf("nices", "since"))
		})
	})
})
