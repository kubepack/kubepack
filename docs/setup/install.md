---
title: Install
menu:
  docs_0.1.0-alpha.0:
    identifier: install-kubepack
    name: Install
    parent: setup
    weight: 10
menu_name: docs_0.1.0-alpha.0
section_menu_id: setup
---

> New to Kubepack? Please start [here](/docs/guides/README.md).

# Installation Guide

## Install Kubepack CLI
Kubepack provides a CLI to work with database objects. Download pre-built binaries from [kubepack/kubepack Github releases](https://github.com/kubepack/pack/releases) and put the binary to some directory in your `PATH`. To install on Linux 64-bit and MacOS 64-bit you can run the following commands:

```console
# Linux amd 64-bit
wget -O pack https://github.com/kubepack/pack/releases/download/0.1.0-alpha.0/kubepack-linux-amd64 \
  && chmod +x pack \
  && sudo mv pack /usr/local/bin

# Mac 64-bit
wget -O pack https://github.com/kubepack/pack/releases/download/0.1.0-alpha.0/kubepack-darwin-amd64 \
  && chmod +x pack \
  && sudo mv pack /usr/local/bin
```

If you prefer to install Kubepack cli from source code, you will need to set up a GO development environment following [these instructions](https://golang.org/doc/code.html). Then, install `kubepack` CLI using `go get` from source code.

```console
go get github.com/kubepack/pack/...
```

Please note that this will install Kubepack cli from master branch which might include breaking and/or undocumented changes.
