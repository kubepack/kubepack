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

package lib

type ScriptOptions struct {
	DisableApplicationCRD bool
	OsIndependentScript   bool
}

type ScriptOption interface {
	Apply(opt *ScriptOptions)
}

type ScriptOptionFunc func(opt *ScriptOptions)

func (fn ScriptOptionFunc) Apply(opt *ScriptOptions) {
	fn(opt)
}

var DisableApplicationCRD = ScriptOptionFunc(func(opt *ScriptOptions) {
	opt.DisableApplicationCRD = true
})

var OsIndependentScript = ScriptOptionFunc(func(opt *ScriptOptions) {
	opt.OsIndependentScript = true
})
