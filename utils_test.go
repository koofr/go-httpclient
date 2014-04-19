package httpclient_test

import (
	. "git.koofr.lan/go-httpclient.git"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("EscapePath", func() {
	It("should escape path", func() {
		Expect(EscapePath("foo+bar baz?&")).To(Equal("foo%2bbar%20baz%3F&"))
	})
})
