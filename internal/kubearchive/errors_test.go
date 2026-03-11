/*
Copyright 2021 Red Hat, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package kubearchive

import (
	"errors"
	"net/http"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("KubeArchive Errors", func() {
	Describe("Err", func() {
		It("implements error interface", func() {
			err := Err{
				code: http.StatusNotFound,
				err:  "resource not found",
			}

			Expect(err.Error()).To(ContainSubstring("resource not found"))
			Expect(err.Error()).To(ContainSubstring("404"))
		})

		It("returns correct HTTP status code", func() {
			err := Err{
				code: http.StatusForbidden,
				err:  "forbidden",
			}

			Expect(err.Code()).To(Equal(http.StatusForbidden))
		})
	})

	Describe("IsNotFound", func() {
		It("returns true for 404 errors", func() {
			err := Err{code: http.StatusNotFound, err: "not found"}
			Expect(IsNotFound(err)).To(BeTrue())
		})

		It("returns false for other errors", func() {
			err := Err{code: http.StatusBadRequest, err: "bad request"}
			Expect(IsNotFound(err)).To(BeFalse())
		})

		It("returns false for non-kubearchive errors", func() {
			err := errors.New("some other error")
			Expect(IsNotFound(err)).To(BeFalse())
		})
	})

	Describe("IsBadRequest", func() {
		It("returns true for 400 errors", func() {
			err := Err{code: http.StatusBadRequest, err: "bad request"}
			Expect(IsBadRequest(err)).To(BeTrue())
		})

		It("returns false for other errors", func() {
			err := Err{code: http.StatusNotFound, err: "not found"}
			Expect(IsBadRequest(err)).To(BeFalse())
		})
	})

	Describe("IsUnauthorized", func() {
		It("returns true for 401 errors", func() {
			err := Err{code: http.StatusUnauthorized, err: "unauthorized"}
			Expect(IsUnauthorized(err)).To(BeTrue())
		})

		It("returns false for other errors", func() {
			err := Err{code: http.StatusForbidden, err: "forbidden"}
			Expect(IsUnauthorized(err)).To(BeFalse())
		})
	})

	Describe("IsForbidden", func() {
		It("returns true for 403 errors", func() {
			err := Err{code: http.StatusForbidden, err: "forbidden"}
			Expect(IsForbidden(err)).To(BeTrue())
		})

		It("returns false for other errors", func() {
			err := Err{code: http.StatusUnauthorized, err: "unauthorized"}
			Expect(IsForbidden(err)).To(BeFalse())
		})
	})

	Describe("IsMultipleResourcesFound", func() {
		It("returns true for 500 errors with specific message", func() {
			err := Err{
				code: http.StatusInternalServerError,
				err:  "more than one resource found",
			}
			Expect(IsMultipleResourcesFound(err)).To(BeTrue())
		})

		It("returns false for 500 errors without specific message", func() {
			err := Err{
				code: http.StatusInternalServerError,
				err:  "internal server error",
			}
			Expect(IsMultipleResourcesFound(err)).To(BeFalse())
		})

		It("returns false for other status codes even with message", func() {
			err := Err{
				code: http.StatusBadRequest,
				err:  "more than one resource found",
			}
			Expect(IsMultipleResourcesFound(err)).To(BeFalse())
		})
	})

	Describe("codeForError", func() {
		It("extracts code from Err type", func() {
			err := Err{code: http.StatusNotFound, err: "not found"}
			Expect(codeForError(err)).To(Equal(http.StatusNotFound))
		})

		It("returns -1 for unknown error types", func() {
			err := errors.New("standard error")
			Expect(codeForError(err)).To(Equal(-1))
		})
	})

	Describe("messageForError", func() {
		It("extracts message from Err type", func() {
			err := Err{code: http.StatusNotFound, err: "resource not found"}
			Expect(messageForError(err)).To(Equal("resource not found"))
		})

		It("returns Error() for unknown error types", func() {
			err := errors.New("standard error")
			Expect(messageForError(err)).To(Equal("standard error"))
		})
	})
})
