/*
Copyright AppsCode Inc. and Contributors

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

package e2e_test

import (
	"kubepack.dev/kubepack/test/e2e/framework"

	. "github.com/onsi/ginkgo"
	_ "github.com/onsi/gomega"
)

var _ = Describe("E2E Tests", func() {
	var (
		f *framework.Invocation
	)
	BeforeEach(func() {
		f = root.Invoke()
	})
	Describe("Test", func() {
		Context("Context", func() {
			It("should complete successfully", func() {
				By("A test step")
				_ = f.CreateNamespace()
				_ = f.DeleteNamespace()
			})
		})
	})
})
