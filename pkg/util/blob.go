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

package util

import (
	"context"
	"io/ioutil"
	"os"
	"strings"

	"gocloud.dev/blob"
	"golang.org/x/oauth2/google"
)

const GoogleApplicationCredentials = "/home/tamal/Downloads/appscode-domains-1577f17c3fd8.json"

func Upload(dir, filename string, data []byte) error {
	saKey, err := ioutil.ReadFile(GoogleApplicationCredentials)
	if err != nil {
		return err
	}

	cfg, err := google.JWTConfigFromJSON(saKey)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(GoogleApplicationCredentials+"-private-key", cfg.PrivateKey, 0644)
	if err != nil {
		return err
	}

	os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", GoogleApplicationCredentials)

	ctx := context.Background()
	bucket, err := blob.OpenBucket(ctx, "gs://kubepack-usercontent?access_id="+cfg.Email+"&private_key_path="+GoogleApplicationCredentials+"-private-key")
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
