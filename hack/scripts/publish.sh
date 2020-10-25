#!/bin/bash

# Copyright The Kubepack Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -xeou pipefail

REPO_ROOT=$(dirname "${BASH_SOURCE[0]}")/../..

pushd $REPO_ROOT

find testdata/charts -maxdepth 1 -mindepth 1 -type d -exec helm package {} \;
mkdir -p testdata/archives
mv *.tgz testdata/archives

helm repo index testdata/archives/ --url https://bundles.kubepack.com

gsutil rsync -d -r testdata/archives gs://bundles.kubepack.com
gsutil acl ch -u AllUsers:R -r gs://bundles.kubepack.com

sleep 10

gcloud compute url-maps invalidate-cdn-cache cdn \
    --project appscode-domains \
    --host bundles.kubepack.com \
    --path "/index.yaml"

popd
