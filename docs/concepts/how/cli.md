---
title: Using Kubepack CLI
menu:
  docs_0.1.0-alpha.2:
    identifier: cli-how
    name: CLI
    parent: how-concepts
    weight: 25
menu_name: docs_0.1.0-alpha.2
section_menu_id: concepts
---

> New to Kubepack? Please start [here](/docs/concepts/README.md).

# Using Kubepack CLI

### How to Get Dependencies

```console
$ pack dep -f .
```
command will get dependencies defined under `items` field in `dependency-list.yaml` file. You can get specific version, branch or revision.
See tutorial of [dependency-list.yaml](/docs/guides/manifest.md). All the dependencies will appear in `manifests/vendor` folder.
You can get verbose output with `--v=10` or `-v 10` flag.

### Edit File from manifests/vendor Folder

```console
$ pack edit -p <filepath>
```

command edit file, exists in `manifests/vendor` folder and generate patch in `manifests/patch` folder.
This patch file-path will be same as `manifests/vendor` folder.

**Note: `filepath`: is relative file path.**

### Combine manifests/vendor and manifests/patch files

```console
$ pack up -f .
```

command combine files from `manifests/patch` and `manifests/vendor` folder. This combination of `manifests/patch` and `manifests/vendor` files appear in `manifests/output` folder. [Read more](/docs/concepts/how/cli.md)

### Validate manifests/output folder

```console
$ pack validate -f .
```

This command will validate the `manifests/output` folder yaml files using `openapi-spec`.
If some file is not a valid yaml then throws errors. `--kube-version` flag is used specify kubernetes version, which you want to validate against.


## Next Steps

- Want to consume apps published using Kubepack? Please visit [here](/docs/concepts/how/user.md).
- To learn about `dependency-list.yaml` file, please visit [here](/docs/concepts/how/manifest.md).
- Learn more about **Pack** jsonnet-support [here](/docs/concepts/how/jsonnet-support.md).
- Learn more about **Pack** validation [here](/docs/concepts/how/validation.md).
