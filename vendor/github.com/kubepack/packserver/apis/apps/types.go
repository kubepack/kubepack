/*
Copyright 2018 The Kubepack Authors.

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

package apps

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// PackList is a list of Pack objects.
type PackList struct {
	metav1.TypeMeta
	metav1.ListMeta

	Items []Pack
}

type PackSpec struct {
}

type PackStatus struct {
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type Pack struct {
	metav1.TypeMeta
	metav1.ObjectMeta

	Spec   PackSpec
	Status PackStatus
}

// +genclient
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type User struct {
	metav1.TypeMeta
	metav1.ObjectMeta

	// DisallowedPacks holds a list of Pack.Names that are disallowed.
	DisallowedPacks []string
}

// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// UserList is a list of User objects.
type UserList struct {
	metav1.TypeMeta
	metav1.ListMeta

	// Items is a list of Users
	Items []User
}
