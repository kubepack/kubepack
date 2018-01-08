#!/usr/bin/env bash

pushd $GOPATH/src/github.com/kubepack/kubepack/hack/gendocs
go run main.go
popd
