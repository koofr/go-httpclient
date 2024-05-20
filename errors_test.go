package httpclient_test

import (
	"fmt"
	"net/http"

	. "github.com/koofr/go-httpclient"

	. "github.com/onsi/ginkgo/v2"
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

	Describe("IsInvalidStatusCode", func() {
		It("should check if status code matches", func() {
			err := InvalidStatusError{
				Expected: []int{200},
				Got:      409,
				Headers:  make(http.Header),
				Content:  "Error",
			}

			var _ error = err

			Expect(IsInvalidStatusCode(err, 409)).To(BeTrue())
		})

		It("should return false if error is not valid", func() {
			Expect(IsInvalidStatusCode(fmt.Errorf("error"), 409)).To(BeFalse())
		})
	})
})
