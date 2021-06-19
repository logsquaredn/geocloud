package worker_test

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/logsquaredn/geocloud/worker"
)

var (
	digest = "sha256:abc123abc123abc123abc123abc123abc123abc123abc123abc123abc123abc1"
	tag = "latest"
	registryurl = "test.io"
	slashimage = "logsquaredn/geocloud"
	noslashimage = "golang"
)

var _ = Describe("Registry", func() {
	registry, err := worker.NewRegistry(registryurl)
	Expect(err).ToNot(HaveOccurred())

	Describe("Images with a '/' in their name", func() {
		Describe("And a digest", func() {
			It("Returns the correct ref", func() {
				ref := registry.Ref(fmt.Sprintf("%s@%s", slashimage, digest))
				Expect(ref).To(Equal(fmt.Sprintf("%s/%s@%s", registryurl, slashimage, digest)))
			})
		})

		Describe("And a tag", func() {
			It("Returns the correct ref", func() {
				ref := registry.Ref(fmt.Sprintf("%s:%s", slashimage, tag))
				Expect(ref).To(Equal(fmt.Sprintf("%s/%s:%s", registryurl, slashimage, tag)))
			})
		})

		Describe("And no tag or digest", func() {
			It("Returns the correct ref", func() {
				ref := registry.Ref(slashimage)
				Expect(ref).To(Equal(fmt.Sprintf("%s/%s:%s", registryurl, slashimage, "latest")))
			})
		})
	})

	Describe("Images without a '/' in their name", func() {
		Describe("And a digest", func() {
			It("Returns the correct ref", func() {
				ref := registry.Ref(fmt.Sprintf("%s@%s", noslashimage, digest))
				Expect(ref).To(Equal(fmt.Sprintf("%s/library/%s@%s", registryurl, noslashimage, digest)))
			})
		})

		Describe("And a tag", func() {
			It("Returns the correct ref", func() {
				ref := registry.Ref(fmt.Sprintf("%s:%s", noslashimage, tag))
				Expect(ref).To(Equal(fmt.Sprintf("%s/library/%s:%s", registryurl, noslashimage, tag)))
			})
		})

		Describe("And no tag or digest", func() {
			It("Returns the correct ref", func() {
				ref := registry.Ref(noslashimage)
				Expect(ref).To(Equal(fmt.Sprintf("%s/library/%s:%s", registryurl, noslashimage, "latest")))
			})
		})
	})
})
