---
title: Scenarios | Kubepack
menu:
  docs_0.1.0-alpha.0:
    identifier: s1-guides
    name: Schenario 1
    parent: guides
    weight: 70
menu_name: docs_0.1.0-alpha.0
section_menu_id: guides
---

> New to Kubepack? Please start [here](/docs/concepts/README.md).

# Scenario-1

**This docs trying to explain the behavior of Pack**
***

This section explain [test-1](https://github.com/kubepack/pack/tree/master/_testdata/test-1).

If you look into this test's `manifest.yaml` file.
```console
$ cat manifest.yaml

package: github.com/kubepack/pack/_testdata/test-1
owners:
- name: Appscode
  email: team@appscode.com
dependencies:
- package: github.com/kubepack/kube-a
  branch: test-1
```
You'll see it depends on repository [kube-a](https://kubepack/kube-a) of branch `test-1`.

You can see the whole dependencies in below image.

![alt text](/_testdata/test-1/test-1.jpg)

### Explanation

1. [test-1](https://github.com/kubepack/pack/tree/master/_testdata/test-1) directly depends on [kube-a](https://kubepack/kube-a) of branch `test-1`.
2. [kube-a](https://kubepack/kube-a) depends on  [kube-b](https://kubepack/kube-b) of branch `test-1`.
`kube-a` contains the patch patch of `kube-b`'s `nginx-deployment.yaml` file.
3. [kube-b](https://kubepack/kube-b) depends on [kube-c](https://kubepack/kube-c) of branch `test-1`.
`kube-b` contains the patch patch of `kube-c`'s `nginx-deployment.yaml` file.

When run `pack dep` in `test-1`, following things happen.

1. Get all the dependencies, reading `manifest.yaml` file.
2. `kube-b`'s `nginx-deployment.yaml` file is combination of patch (exists in `kube-a` repository) and original file (exists in `kube-b` repository).
3. `kube-c`'s `nginx-deployment.yaml` file is combination of patch (exists in `kube-b` repository) and original file (exists in `kube-c` repository).

## Next Steps

- Want to publish apps using Kubepack? Please visit [here](/docs/concepts/how/publisher.md).
- Want to consume apps published using Kubepack? Please visit [here](/docs/concepts/how/user.md).
- To learn about `manifest.yaml` file, please visit [here](/docs/concepts/how/manifest.md).
- Learn more about `pack` cli from [here](/docs/concepts/how/cli.md).
