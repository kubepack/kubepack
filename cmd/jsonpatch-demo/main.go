/*
Copyright The Kubepack Authors.

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

package main

import (
	"fmt"

	jsonpatch "github.com/evanphx/json-patch"
)

func main() {
	p2, err := jsonpatch.DecodePatch([]byte(`[{ "op": "replace", "path": "/licenseKey", "value": "xyz" }]`))
	if err != nil {
		panic(err)
	}

	d2, err := p2.ApplyIndent([]byte(`{ "a": "b" }`), "  ")
	if err != nil {
		panic(err)
	}
	fmt.Println(string(d2))
}