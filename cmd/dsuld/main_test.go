package main_test

import (
	"os/exec"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("Dsuld", func() {
	var session *gexec.Session

	BeforeEach(func() {
		// ...
	})

	AfterEach(func() {
		gexec.CleanupBuildArtifacts()
	})

	Describe("Calling dsuld with arguments", func() {
		dsuldPath := buildDsuld()

		Context("ask for version", func() {
			session = runDsuld(dsuldPath, "-v")

			It("prints 'dsuld v0.0.0' to stdout", func() {
				Eventually(session).Should(gbytes.Say("dsuld v0.0.0"))

			})
			It("exits with status code 0", func() {
				Eventually(session).Should(gexec.Exit(0))
			})
		})
	})
})

func buildDsuld() string {
	dsulPath, err := gexec.Build("github.com/hymnis/dsul-go/cmd/dsuld")
	Expect(err).NotTo(HaveOccurred())

	return dsulPath
}

func runDsuld(path string, args string) *gexec.Session {
	cmd := exec.Command(path, args)
	session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
	Expect(err).NotTo(HaveOccurred())

	return session
}
