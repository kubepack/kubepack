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

import (
	"kubepack.dev/kubepack/apis"

	"gomodules.xyz/blobfs"
	"gomodules.xyz/blobfs/testing"
)

const GoogleApplicationCredentials = "/personal/AppsCode/credentials/bytebuilders@byte-builders.json"

type BlobStore struct {
	Host   string
	Bucket string
	blobfs.Interface
}

func NewTestBlobStore() (*BlobStore, error) {
	fs, err := testing.NewTestGCS(apis.YAMLBucket, GoogleApplicationCredentials)
	if err != nil {
		return nil, err
	}
	return &BlobStore{
		Interface: fs,
		Host:   apis.YAMLHost,
		Bucket: apis.YAMLBucket,
	}, nil
}
