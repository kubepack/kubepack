---
title: Developer Guide
menu:
  docs_0.1.0-alpha.1:
    identifier: developer-guide-readme
    name: Overview
    parent: developer-guide
    weight: 10
menu_name: docs_0.1.0-alpha.1
section_menu_id: setup
aliases:
  - /docs/0.1.0-alpha.1/setup/developer-guide/
---

# Development Guide

This document is intended to be the canonical source of truth for things like supported toolchain versions for building Pack.
If you find a requirement that this doc does not capture, please submit an issue on github.

This document is intended to be relative to the branch in which it is found. It is guaranteed that requirements will change over time for the development branch, but release branches of Pack should not change.

### Build Pack
Some of the Pack development helper scripts rely on a fairly up-to-date GNU tools environment, so most recent Linux distros should work just fine out-of-the-box.

#### Setup GO
Pack is written in Google's GO programming language. Currently, Pack is developed and tested on **go 1.9.2**. If you haven't set up a GO development environment, please follow [these instructions](https://golang.org/doc/code.html) to install GO.

#### Download Source

```console
$ go get github.com/kubepack/pack
$ cd $(go env GOPATH)/src/github.com/kubepack/pack
```

#### Install Dev tools
To install various dev tools for Pack, run the following command:

```console
# setting up dependencies for compiling pack...
$ ./hack/builddeps.sh
```

#### Build Binary
```
$ ./hack/make.py
$ pack version
```

#### Dependency management
Pack uses [Dep](https://github.com/golang/dep) to manage dependencies. Dependencies are already checked in the `vendor` folder.
If you want to update/add dependencies, run:
```console
$ dep ensure
```

#### Generate Kubepack Reference Docs
```console
$ ./hack/gendocs/make.sh
```
