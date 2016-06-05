package match_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestMatch(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Match Suite")
}
