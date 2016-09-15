package config_test

import (
	"io/ioutil"
	"os"

	. "github.com/iancmcc/jig/config"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Jig root", func() {

	var tempdir string

	BeforeEach(func() {
		os.Setenv("JIGROOT", "")
		td, err := ioutil.TempDir("", "jig-")
		if err != nil {
			panic(err)
		}
		tempdir = td
	})

	AfterEach(func() {
		if tempdir != "" {
			os.RemoveAll(tempdir)
		}
		tempdir = ""
	})

	It("should be found when the dir exists", func() {
		Expect(CreateJigRoot(tempdir)).To(BeNil())
		Expect(IsJigRoot(tempdir)).To(Equal(true))
		Expect(FindClosestJigRoot(tempdir)).To(Equal(tempdir))
	})

	It("should throw an error when the dir does not exist", func() {
		Expect(IsJigRoot(tempdir)).To(Equal(false))
		root, err := FindClosestJigRoot(tempdir)
		Expect(root).To(BeEmpty())
		Expect(err).To(Not(BeNil()))
	})

})
