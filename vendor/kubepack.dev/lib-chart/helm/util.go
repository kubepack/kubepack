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

package helm

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"helm.sh/helm/v3/pkg/helmpath/xdg"
	authv1 "k8s.io/api/authorization/v1"
	authv1client "k8s.io/client-go/kubernetes/typed/authorization/v1"
	"k8s.io/client-go/rest"
)

const yamlSeparator = "\n---"

func isSARAllowed(sar authv1.SelfSubjectAccessReview, config *rest.Config) (bool, error) {
	client, err := authv1client.NewForConfig(config)
	if err != nil {
		return false, err
	}

	result, err := client.SelfSubjectAccessReviews().Create(&sar)
	if err != nil {
		return false, err
	}

	return result.Status.Allowed, nil
}

func splitYAML(data []byte, output *[]string) {
	sep := len([]byte(yamlSeparator))
	if i := bytes.Index(data, []byte(yamlSeparator)); i >= 0 {
		*output = append(*output, string(data[0:i]))
		splitYAML(data[i+sep:], output)
	} else {
		*output = append(*output, string(data))
	}
}

func debug(format string, v ...interface{}) {
	format = fmt.Sprintf("[debug] %s\n", format)
	log.Output(2, fmt.Sprintf(format, v...))
}

func setEnv(chartDir string) error {
	err := os.Setenv(xdg.CacheHomeEnvVar, filepath.Join(chartDir, "cache"))
	if err != nil {
		return err
	}

	err = os.Setenv(xdg.ConfigHomeEnvVar, filepath.Join(chartDir, "config"))
	if err != nil {
		return err
	}

	return os.Setenv(xdg.DataHomeEnvVar, filepath.Join(chartDir, "data"))
}

func unsetEnv() error {
	err := os.Unsetenv(xdg.CacheHomeEnvVar)
	if err != nil {
		return err
	}

	err = os.Unsetenv(xdg.ConfigHomeEnvVar)
	if err != nil {
		return err
	}

	return os.Unsetenv(xdg.DataHomeEnvVar)
}
