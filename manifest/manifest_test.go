package manifest_test

import (
	"strings"

	. "github.com/iancmcc/jig/manifest"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var (
	json = strings.TrimSpace(`
	[{
		"repo": "github.com/iancmcc/jig",
		"ref": "develop"
	},{
		"repo": "github.com/zenoss/zenoss",
		"ref": "master"
	}]
	`)
)

var _ = Describe("Manifest from JSON", func() {
	It("should deserialize from JSON", func() {
		results, err := FromJSON(strings.NewReader(json))
		Expect(err).To(BeNil())
		Expect(results).To(Not(BeNil()))
		Expect(results.Repos).To(HaveLen(2))
	})
})
