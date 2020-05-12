[![Go Report Card](https://goreportcard.com/badge/kubepack.dev/kubepack)](https://goreportcard.com/report/kubepack.dev/kubepack)
[![Build Status](https://github.com/kubepack/kubepack/workflows/CI/badge.svg)](https://github.com/kubepack/kubepack/actions?workflow=CI)
[![codecov](https://codecov.io/gh/kubepack/kubepack/branch/master/graph/badge.svg)](https://codecov.io/gh/kubepack/kubepack)
[![Docker Pulls](https://img.shields.io/docker/pulls/kubepack/kubepack-operator.svg)](https://hub.docker.com/r/kubepack/kubepack-operator/)
[![Slack](http://slack.kubernetes.io/badge.svg)](http://slack.kubernetes.io/#kubepack)
[![Twitter](https://img.shields.io/twitter/follow/kubepack.svg?style=social&logo=twitter&label=Follow)](https://twitter.com/intent/follow?screen_name=Kubepack)

# Kubepack

## Configure Helm

```console
helm repo add kubepack-bundles https://bundles.kubepack.com
helm repo add stable https://kubernetes-charts.storage.googleapis.com
helm repo update
```

## Generate Test Bundles

```console

$ go run cmd/bundle-generator/main.go --name=csi-vault-bundle --charts https://charts.appscode.com/stable/@csi-vault@v0.3.0

$ go run cmd/bundle-generator/main.go --name=vault-operator-bundle \
  --charts https://charts.appscode.com/stable/@vault-operator@v0.3.0 \
  --charts https://charts.appscode.com/stable/@vault-catalog@v0.3.0 \
  --bundles https://bundles.kubepack.com@csi-vault-bundle@v0.3.0

$ go run cmd/bundle-generator/main.go --name=stash-mongodb-bundle \
  --charts https://charts.appscode.com/stable/@stash-mongodb@3.4.22:true,3.6.13:true,4.0.11:true,4.1.13:true@required@anyof

$ go run cmd/bundle-generator/main.go --name=stash-bundle \
  --charts https://bundles.kubepack.com@stash@v0.9.0-rc.6 \
  --bundles https://bundles.kubepack.com@stash-mongodb-bundle@v0.9.0-rc.6

# $ go run cmd/bundle-generator/main.go --name=cert-manager-bundle --charts https://charts.jetstack.io@cert-manager@v0.12.0@@@cert-manager

$ go run cmd/bundle-generator/main.go --name=kubedb-bundle \
  --charts https://charts.appscode.com/stable/@kubedb@v0.13.0-rc.0 \
  --charts https://charts.appscode.com/stable/@kubedb-catalog@v0.13.0-rc.0 \
  --charts https://charts.jetstack.io@cert-manager@v0.12.0@optional@@cert-manager \
  --bundles https://bundles.kubepack.com@stash-bundle@v0.9.0-rc.6

$ ./hack/publish-testcharts.sh
```

## Generate BundleView from a Bundle

```console
$ go run cmd/bundleview-generator/main.go
```

## Generate Order from a BundleView

```console
$ go run cmd/order-generator/main.go
```

## Generate Application from a selected Chart in an Order

```console
$ go run cmd/app-generator/main.go
```

## Generate PackageView for a Chart

```console
$ go run cmd/packageview-generator/main.go
```

## Generate Install scripts

**Site for Hosting User YAMLs & Scripts**

[https://usercontent.kubepack.com](https://usercontent.kubepack.com). These files are *public* and hosted on Google Cloud Storage Bucket `gs://kubepack-usercontent`.

**Helm 3**
```console
$ go run cmd/helm3-command-generator/main.go
```

**Helm 2**
```console
$ go run cmd/helm2-command-generator/main.go
```

**YAML**
```console
$ go run cmd/install-yaml-generator/main.go
```

**Check Permission**
```console
$ go run cmd/permission-checker/main.go
```

**Install / Uninstall Chart**
```console
$ go run cmd/install-order/main.go
$ go run cmd/uninstall-order/main.go
```

## Read Helm Hub index to determine Chart Repository Name

```console
$ go run cmd/helm-hub-reader/main.go
```

## API Server

- http://localhost:4000/products
- http://localhost:4000/products/appscode/kubedb
- http://localhost:4000/product_id/prod_Gnc33bJka9iRl9
- http://localhost:4000/products/appscode/kubedb/plans
- http://localhost:4000/products/appscode/kubedb/compare
- http://localhost:4000/products/appscode/kubedb/plans/kubedb-community
- http://localhost:4000/products/appscode/kubedb/plans/kubedb-community/bundleview

### Generate PackageView

- http://localhost:4000/packageview?url=https://charts.appscode.com/stable/&name=kubedb&version=v0.13.0-rc.0
- http://localhost:4000/packageview?url=https://bundles.kubepack.com&name=stash&version=v0.9.0-rc.6

### Generate BundleView for Chart

- http://localhost:4000/bundleview?url=https://charts.appscode.com/stable/&name=kubedb&version=v0.13.0-rc.0

### Generate order

```console
curl -X POST -H "Content-Type: application/json" -d @artifacts/kubedb-community/bundleview.json http://localhost:4000/deploy/orders
```

```json
{"kind":"Order","apiVersion":"kubepack.com/v1alpha1","metadata":{"name":"kubedb-community","uid":"1f1d149b-5226-4659-8feb-165face489b3","creationTimestamp":"2020-02-26T12:00:24Z"},"spec":{"items":[{"chart":{"url":"https://charts.appscode.com/stable/","name":"kubedb","version":"v0.13.0-rc.0","releaseName":"kubedb","namespace":"kube-system","bundle":{"name":"kubedb-community","url":"https://bundles.kubepack.com","version":"v0.13.0-rc.0"}}},{"chart":{"url":"https://charts.appscode.com/stable/","name":"kubedb-catalog","version":"v0.13.0-rc.0","releaseName":"kubedb-catalog","namespace":"kube-system","bundle":{"name":"kubedb-community","url":"https://bundles.kubepack.com","version":"v0.13.0-rc.0"}}}]},"status":{}}
```

- http://localhost:4000/deploy/orders/1f1d149b-5226-4659-8feb-165face489b3/helm2
- http://localhost:4000/deploy/orders/1f1d149b-5226-4659-8feb-165face489b3/helm3
- http://localhost:4000/deploy/orders/1f1d149b-5226-4659-8feb-165face489b3/yaml

### List Helm Hub repositories, Charts and Chart Versions

- http://localhost:4000/chartrepositories
- http://localhost:4000/chartrepositories/charts/?url=https://charts.appscode.com/stable/
- http://localhost:4000/chartrepositories/charts/voyager/versions/?url=https://charts.appscode.com/stable/
