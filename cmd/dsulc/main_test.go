package main_test

import (
	"os/exec"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("Dsulc", func() {
	var session *gexec.Session

	BeforeEach(func() {
		// ...
	})

	AfterEach(func() {
		gexec.CleanupBuildArtifacts()
	})

	Describe("Calling dsulc with arguments", func() {
		dsulcPath := buildDsulc()

		Context("ask for version", func() {
			session = runDsulc(dsulcPath, "-v")

			It("prints 'dsulc v0.0.0' to stdout", func() {
				Eventually(session).Should(gbytes.Say("dsulc v0.0.0"))

			})
			It("exits with status code 0", func() {
				Eventually(session).Should(gexec.Exit(0))
			})
		})

	})
})

func buildDsulc() string {
	dsulPath, err := gexec.Build("github.com/hymnis/dsul-go/cmd/dsulc")
	Expect(err).NotTo(HaveOccurred())

	return dsulPath
}

func runDsulc(path string, args string) *gexec.Session {
	cmd := exec.Command(path, args)
	session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
	Expect(err).NotTo(HaveOccurred())

	return session
}
