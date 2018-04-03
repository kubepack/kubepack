---
title: Scenarios | Kubepack
menu:
  docs_0.1.0-alpha.2:
    identifier: s11-guides
    name: Scenario 11
    parent: guides
    weight: 75
menu_name: docs_0.1.0-alpha.2
section_menu_id: guides
---

> New to Kubepack? Please start [here](/docs/concepts/README.md).

# Scenario-11

**This doc explains how `pack up` command works and how install.sh file is generated.**

![alt text](/docs/_testdata/test-11/test-11.jpg)

To learn how this works, clone [this repository](https://kubepack/kube-a) and checkout branch `test-11`. You will find `dependency-list.yaml` file in the project root.

```console
$ cat dependency-list.yaml
items:
  - package: github.com/kubepack/kube-b
    branch: test-11
  - package: github.com/kubepack/kube-c
    branch: test-11
```

Now, run `pack up -f .` command in this project root. This executes the following steps.

 - Combines `manifests/vendor` and `manifests/patch` to `manifests/output` folder.
 - Generates a DAG(Directed Acyclic Graph) from `dependency-list.yaml`. From this dependency graph, it generates a `install.sh` file. This installer script contains commands to deploy `manifests/output` folder. Each parent package can provide their own `install.sh` script. If no script is provided, `kubectl apply -R -f .` command will be used to install a package.

This is what the auto-generated installer script looks like:

```console
cat manifests/output/install.sh

#!/bin/bash

pushd manifests/output/github.com/kubepack/kube-b
kubectl apply -R -f .
popd

pushd manifests/output/github.com/kubepack/kube-c
kubectl apply -R -f .
popd

pushd manifests/output/github.com/kubepack/kube-d
kubectl apply -R -f .
popd

pushd manifests/output/github.com/kubepack/kube-e
kubectl apply -R -f .
popd

pushd manifests/output/github.com/kubepack/kube-f
kubectl apply -R -f .
popd

pushd manifests/output/github.com/kubepack/kube-a
kubectl apply -R -f .
popd
```

 - At first there will be `kubectl apply` command for `kube-b` or `kube-c`, whichever comes first in `dependency-list.yaml`.
 As `kube-b` appears first in `dependency-list.yaml`, `kube-b` will be first, `kube-c` second.

 - `kube-b` will be process first. That's why `kube-d` will appear before `kube-e`.
 After `kube-d`, `kube-e` will appear.

 - Then `kube-f` and at last, root repo.

If any repository's `manifests/app` folder contain an `install.sh` file, then it will be used instead. Users can use their customize commands for deploy, these customize commands should be in `manifests/app/install.sh` file of that repository.


## Next Steps

- Want to publish apps using Kubepack? Please visit [here](/docs/concepts/how/publisher.md).
- Want to consume apps published using Kubepack? Please visit [here](/docs/concepts/how/user.md).
- To learn about `dependency-list.yaml` file, please visit [here](/docs/concepts/how/manifest.md).
- Learn more about `pack` cli from [here](/docs/concepts/how/cli.md).