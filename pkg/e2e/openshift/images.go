// Package openshift runs the OpenShift extended test suite.
package openshift

import (
	"github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/openshift/osde2e/pkg/common/alert"
	"github.com/openshift/osde2e/pkg/common/helper"
)

var _ = ginkgo.Describe("[Suite: openshift][image-registry]", func() {
	defer ginkgo.GinkgoRecover()
	ginkgo.BeforeEach(func() {
		alert.RegisterGinkgoAlert(ginkgo.CurrentGinkgoTestDescription().TestText, "SD-CICD", "Jeffrey Sica", "sd-cicd-alerts", "sd-cicd@redhat.com", 4)
	})
	h := helper.New()

	e2eTimeoutInSeconds := 3600
	ginkgo.It("should run until completion", func() {
		// configure tests
		cfg := DefaultE2EConfig
		cfg.Suite = "openshift/image-registry"
		cmd := cfg.Cmd()

		// setup runner
		r := h.Runner(cmd)

		r.Name = "openshift-image-registry"

		// run tests
		stopCh := make(chan struct{})
		err := r.Run(e2eTimeoutInSeconds, stopCh)
		Expect(err).NotTo(HaveOccurred())

		// get results
		results, err := r.RetrieveResults()
		Expect(err).NotTo(HaveOccurred())

		// write results
		h.WriteResults(results)
	}, float64(e2eTimeoutInSeconds+30))
})

var _ = ginkgo.Describe("[Suite: openshift][image-ecosystem]", func() {
	defer ginkgo.GinkgoRecover()
	ginkgo.BeforeEach(func() {
		alert.RegisterGinkgoAlert(ginkgo.CurrentGinkgoTestDescription().TestText, "SD-CICD", "Jeffrey Sica", "sd-cicd-alerts", "sd-cicd@redhat.com", 4)
	})
	h := helper.New()

	e2eTimeoutInSeconds := 3600
	ginkgo.It("should run until completion", func() {
		// configure tests
		cfg := DefaultE2EConfig
		cfg.Suite = "openshift/image-ecosystem"
		cmd := cfg.Cmd()

		// setup runner
		r := h.Runner(cmd)

		r.Name = "openshift-image-ecosystem"

		// run tests
		stopCh := make(chan struct{})
		err := r.Run(e2eTimeoutInSeconds, stopCh)
		Expect(err).NotTo(HaveOccurred())

		// get results
		results, err := r.RetrieveResults()
		Expect(err).NotTo(HaveOccurred())

		// write results
		h.WriteResults(results)
	}, float64(e2eTimeoutInSeconds+30))
})
