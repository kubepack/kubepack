[![Go Report Card](https://goreportcard.com/badge/kubepack.dev/kubepack)](https://goreportcard.com/report/kubepack.dev/kubepack)
[![Build Status](https://github.com/kubepack/kubepack/workflows/CI/badge.svg)](https://github.com/kubepack/kubepack/actions?workflow=CI)
[![codecov](https://codecov.io/gh/kubepack/kubepack/branch/master/graph/badge.svg)](https://codecov.io/gh/kubepack/kubepack)
[![Docker Pulls](https://img.shields.io/docker/pulls/kubepack/kubepack-operator.svg)](https://hub.docker.com/r/kubepack/kubepack-operator/)
[![Slack](http://slack.kubernetes.io/badge.svg)](http://slack.kubernetes.io/#kubepack)
[![Twitter](https://img.shields.io/twitter/follow/kubepack.svg?style=social&logo=twitter&label=Follow)](https://twitter.com/intent/follow?screen_name=Kubepack)

# Kubepack

## Configure Helm

```console
helm repo add kubepack-testcharts https://kubepack-testcharts.storage.googleapis.com
helm repo add stable https://kubernetes-charts.storage.googleapis.com
helm repo update
```

## Generate Test Bundles

```console

$ go run cmd/bundle-generator/main.go --name=csi-vault-bundle --charts https://charts.appscode.com/stable/@csi-vault@v0.3.0

$ go run cmd/bundle-generator/main.go --name=vault-operator-bundle \
  --charts https://charts.appscode.com/stable/@vault-operator@v0.3.0 \
  --charts https://charts.appscode.com/stable/@vault-catalog@v0.3.0 \
  --bundles https://kubepack-testcharts.storage.googleapis.com@csi-vault-bundle@v0.3.0

$ go run cmd/bundle-generator/main.go --name=stash-mongodb-bundle \
  --charts https://charts.appscode.com/stable/@stash-mongodb@3.4.22:true,3.6.13:true,4.0.11:true,4.1.13:true@required@anyof

$ go run cmd/bundle-generator/main.go --name=stash-bundle \
  --charts https://charts.appscode.com/stable/@stash@v0.9.0-rc.2 \
  --bundles https://kubepack-testcharts.storage.googleapis.com@stash-mongodb-bundle@v0.9.0-rc.2

# $ go run cmd/bundle-generator/main.go --name=cert-manager-bundle --charts https://charts.jetstack.io@cert-manager@v0.12.0@@@cert-manager

$ go run cmd/bundle-generator/main.go --name=kubedb-bundle \
  --charts https://charts.appscode.com/stable/@kubedb@v0.9.0-rc.2 \
  --charts https://charts.appscode.com/stable/@kubedb-catalog@v0.9.0-rc.2 \
  --charts https://charts.jetstack.io@cert-manager@v0.12.0@optional@@cert-manager \
  --bundles https://kubepack-testcharts.storage.googleapis.com@stash-bundle@v0.9.0-rc.2

$ ./hack/publish-testcharts.sh
```
