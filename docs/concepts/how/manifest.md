---
title: Kubepack Manifest
menu:
  docs_0.1.0-alpha.0:
    identifier: manifest-how
    name: Manifest
    parent: how-concepts
    weight: 20
menu_name: docs_0.1.0-alpha.0
section_menu_id: concepts
---

> New to Kubepack? Please start [here](/docs/concepts/README.md).

# Kubepack Manifest

### manifest.yaml

`manifest.yaml` is metadata file for pack like [dep](https://github.com/golang/dep)'s Gopkg.toml and [glide](https://github.com/Masterminds/glide)'s glide.yaml.

 The `manifest.yaml` contains below element:

 1. It names the current package.
 2. It declares the external dependencies

 A brief `manifest.yaml` file looks like this:

 ```yaml
package: github.com/kubepack/pack
owners:
- name: AppsCode
  email: team@appscode.com
dependencies:
- package: github.com/kubepack/kube-a
- package: github.com/kubepack/kube-b
  version: ^1.2.0
  fork:    git@github.com:kubepack/kube-c
- package: github.com/kubepack/kube-d
  branch: test-1
- package: github.com/kubepack/kube-e
  revison: 443d58e40a195d826b8f1fd91f8b5a54653c2f3d
```

  - package: The top level package is the location in the `GOPATH`.
  This is used for things such as making sure an import isn't also importing the top level package.
  - owners: The owners is a list of one or more owners for the project. This can be a person or organization and is useful for things like notifying the owners of a security issue without filing a public bug.
  - dependencies: A list of external package needs to import. Each package can include:
    - package: The name of the package to import and the only non-optional item.
    - version: A semantic version, semantic version range, branch, tag, or commit id to use.
    - branch: A git repository branch name.
    - fork: If the package name isn't the repo location or this is a private repository it can go here. The package will be checked out from the repo and put where the package name specifies. This allows using forks.
    - revision: A commit hash number of the package.
