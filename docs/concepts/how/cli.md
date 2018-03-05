---
title: Using Kubepack CLI
menu:
  docs_0.1.0-alpha.1:
    identifier: cli-how
    name: CLI
    parent: how-concepts
    weight: 25
menu_name: docs_0.1.0-alpha.1
section_menu_id: concepts
---

> New to Kubepack? Please start [here](/docs/concepts/README.md).

# Using Kubepack CLI

### How to Get Dependencies

```console
    $ kubectl plugin pack dep
```
command will get dependencies defined under `dependencies` field in `manifest.yaml` file. You can get specific version, branch or revision.
See tutorial of [manifest.yaml](/docs/guides/manifest.md). All the dependencies will appear in `manifests/vendor` folder.
You can get verbose output with `--v=10` or `-v 10` flag.

### Edit File from manifests/vendor Folder
```console
    $ kubectl plugin pack edit -s <filepath>
```
command edit file, exists in `manifests/vendor` folder and generate patch in `manifests/patch` folder.
This patch file-path will be same as `manifests/vendor` folder.

**Note: `filepath`: is relative file path.**

### Combine manifests/vendor and manifests/patch files

```console
    $ kubectl plugin pack up
```
command combine files from `manifests/patch` and `manifests/vendor` folder. This combination of `manifests/patch` and `manifests/vendor` files appear in `manifests/output` folder.

### Validate manifests/output folder

```console
    $ kubepack validate
```
This command will validate the `manifests/output` folder yaml files using `openapi-spec`.
If some file is not a valid yaml then throws errors. `--kube-version` flag is used specify kubernetes version, which you want to validate against.