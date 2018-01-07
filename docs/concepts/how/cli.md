---
title: Using Kubepack CLI
menu:
  docs_0.1.0-alpha.0:
    identifier: cli-how
    name: CLI
    parent: how-concepts
    weight: 25
menu_name: docs_0.1.0-alpha.0
section_menu_id: concepts
---

> New to Kubepack? Please start [here](/docs/concepts/README.md).

# Using Kubepack CLI

### How to Get Dependencies

```console
    $ pack dep
```
command will get dependencies defined under `dependencies` field in `manifest.yaml` file. You can get specific version, branch or revision.
See tutorial of [manifest.yaml](/docs/guides/manifest.md). All the dependencies will appear in `_vendor` folder.
You can get verbose output with `--v=10` or `-v 10` flag.

### Edit File from _vendor Folder
```console
    $ pack edit -s <filepath>
```
command edit file, exists in `_vendor` folder and generate patch in `patch` folder.
This patch file-path will be same as `_vendor` folder.

**Note: `filepath`: is relative file path.**

### Combine _vendor and patch files

```console
    $ pack up
```
command combine files from `patch` and `_vendor` folder. This combination of `patch` and `_vendor` files appear in `outlook` folder.
