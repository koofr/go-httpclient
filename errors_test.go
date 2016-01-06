package httpclient_test

import (
	. "github.com/koofr/go-httpclient"
	"net/http"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("InvalidStatusError", func() {
	Describe("IsInvalidStatusError", func() {
		It("should check if value is InvalidStatusError", func() {
			err := InvalidStatusError{
				Expected: []int{200},
				Got:      409,
				Headers:  make(http.Header),
				Content:  "Error",
			}

			var _ error = err

			_, ok := IsInvalidStatusError(err)
			Expect(ok).To(BeTrue())
		})

		It("should check if pointer is InvalidStatusError", func() {
			err := &InvalidStatusError{
				Expected: []int{200},
				Got:      409,
				Headers:  make(http.Header),
				Content:  "Error",
			}

			var _ error = err

			_, ok := IsInvalidStatusError(err)
			Expect(ok).To(BeTrue())
		})
	})
})
