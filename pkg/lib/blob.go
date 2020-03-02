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

package lib

import (
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"os"
	"strings"

	"gocloud.dev/blob"
	"golang.org/x/oauth2/google"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
)

const YAMLHost = "https://usercontent.kubepack.com"
const YAMLBucket = "gs://kubepack-usercontent"
const GoogleApplicationCredentials = "/home/tamal/Downloads/appscode-domains-1577f17c3fd8.json"

type BlobStore struct {
	URL    string
	Host   string
	Bucket string
}

func NewBlobStore(url, host, bucket string) *BlobStore {
	return &BlobStore{
		URL:    url,
		Host:   host,
		Bucket: bucket,
	}
}

func NewTestBlobStore() (*BlobStore, error) {
	credential := GoogleApplicationCredentials
	if v, ok := os.LookupEnv("GOOGLE_APPLICATION_CREDENTIALS"); ok {
		credential = v
	} else {
		utilruntime.Must(os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", credential))
	}

	saKey, err := ioutil.ReadFile(credential)
	if err != nil {
		return nil, err
	}

	cfg, err := google.JWTConfigFromJSON(saKey)
	if err != nil {
		return nil, err
	}

	err = ioutil.WriteFile(credential+"-private-key", cfg.PrivateKey, 0644)
	if err != nil {
		return nil, err
	}

	return &BlobStore{
		URL:    "gs://kubepack-usercontent?access_id=" + cfg.Email + "&private_key_path=" + credential + "-private-key",
		Host:   YAMLHost,
		Bucket: YAMLBucket,
	}, nil
}

func (b BlobStore) Upload(ctx context.Context, dir, filename string, data []byte) error {
	bucket, err := blob.OpenBucket(ctx, b.URL)
	if err != nil {
		return err
	}
	bucket = blob.PrefixedBucket(bucket, strings.Trim(dir, "/")+"/")
	defer bucket.Close()

	w, err := bucket.NewWriter(ctx, filename, nil)
	if err != nil {
		return err
	}
	_, writeErr := w.Write(data)
	// Always check the return value of Close when writing.
	closeErr := w.Close()
	if writeErr != nil {
		return writeErr
	}
	if closeErr != nil {
		return closeErr
	}
	return nil
}

func (b BlobStore) Download(ctx context.Context, dir, filename string) ([]byte, error) {
	bucket, err := blob.OpenBucket(ctx, b.URL)
	if err != nil {
		return nil, err
	}
	bucket = blob.PrefixedBucket(bucket, strings.Trim(dir, "/")+"/")
	defer bucket.Close()

	// Open the key "foo.txt" for reading with the default options.
	r, err := bucket.NewReader(ctx, filename, nil)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
