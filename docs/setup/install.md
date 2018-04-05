---
title: Install
menu:
  docs_0.1.0:
    identifier: install-kubepack
    name: Install
    parent: setup
    weight: 10
menu_name: docs_0.1.0
section_menu_id: setup
---

> New to Kubepack? Please start [here](/docs/concepts/README.md).

# Installation Guide

## Install Kubepack CLI
Kubepack provides a CLI to work with Kubernetes objects. Download pre-built binaries from [kubepack/pack Github releases](https://github.com/kubepack/pack/releases) and install the binary as a [`kubectl` plugin](https://kubernetes.io/docs/tasks/extend-kubectl/kubectl-plugins/). To install on Linux 64-bit and MacOS 64-bit you can run the following commands:

```console
# Linux amd 64-bit
wget -O pack https://github.com/kubepack/pack/releases/download/0.1.0/pack-linux-amd64 \
  && chmod +x pack \
  && sudo mv pack /usr/local/bin/

# Mac 64-bit
wget -O pack https://github.com/kubepack/pack/releases/download/0.1.0/pack-darwin-amd64 \
  && chmod +x pack \
  && sudo mv pack /usr/local/bin/
```

If you prefer to install Kubepack cli from source code, you will need to set up a GO development environment following [these instructions](https://golang.org/doc/code.html). Then, install `kubepack` CLI using `go get` from source code.

```console
go get -u github.com/kubepack/pack && pack install
```

Please note that this will install Kubepack cli from master branch which might include breaking and/or undocumented changes.
