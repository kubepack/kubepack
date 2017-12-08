#!/usr/bin/env bash

pushd $GOPATH/src/github.com/kubepack/pack/hack/gendocs
go run main.go
popd
