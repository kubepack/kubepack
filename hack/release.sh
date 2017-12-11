#!/bin/bash
set -xeou pipefail

GOPATH=$(go env GOPATH)
REPO_ROOT="$GOPATH/src/github.com/kubepack/pack"

export APPSCODE_ENV=prod

pushd $REPO_ROOT

rm -rf dist

./hack/make.py build
./hack/make.py push
./hack/make.py update_registry

rm dist/.tag

popd
