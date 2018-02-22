---
title: Scenarios | Kubepack
menu:
  docs_0.1.0-alpha.0:
    identifier: s2-guides
    name: Scenario 2
    parent: guides
    weight: 70
menu_name: docs_0.1.0-alpha.0
section_menu_id: guides
---

> New to Kubepack? Please start [here](/docs/concepts/README.md).

# Scenario-2

**This docs trying to explain the behavior of Pack**
***

This section explain [test-2](https://github.com/kubepack/kubepack/tree/master/docs/_testdata/test-2).

If you look into this test's `manifest.yaml` file.
```console
$ cat manifest.yaml

package: github.com/kubepack/kubepack/docs/_testdata/test-2
owners:
- name: Appscode
  email: team@appscode.com
dependencies:
- package: github.com/kubepack/kube-a
  branch: test-2

```
You'll see it depends on branch `test-2` of repository [kube-a](https://kubepack/kube-a).

You can see the whole dependencies in below image.

![alt text](/docs/_testdata/test-2/test-2.jpg)

### Explanation

1. [test-2](https://github.com/kubepack/kubepack/tree/master/docs/_testdata/test-2) directly depends on branch `test-2` of [kube-a](https://github.com/kubepack/kube-a/blob/test-2/manifest.yaml).

2. [kube-a](https://github.com/kubepack/kube-a/tree/test-2) depends on  [kube-b](https://github.com/kubepack/kube-b/blob/test-2/manifest.yaml) of branch `test-2`. `kube-a` contains the patch patch of `kube-b`'s `nginx-deployment.yaml` file.

3. [kube-b](https://github.com/kubepack/kube-b/tree/test-2) depends on branch `test-2` of [kube-c](https://github.com/kubepack/kube-c/blob/test-2/manifest.yaml). `kube-b` contains the patch patch of `kube-c`'s `nginx-deployment.yaml` file.

When run `$ pack dep` in `test-2`, following things happen.

1. Get all the dependencies, reading `manifest.yaml` file.
2. `kube-b`'s `nginx-deployment.yaml` file is combination of patch (exists in `kube-a` repository) and original file (exists in `kube-b` repository).
3. `kube-c`'s `nginx-deployment.yaml` file is combination of patch (exists in `kube-b` repository) and original file (exists in `kube-c` repository).

## Next Steps

- Want to publish apps using Kubepack? Please visit [here](/docs/concepts/how/publisher.md).
- Want to consume apps published using Kubepack? Please visit [here](/docs/concepts/how/user.md).
- To learn about `manifest.yaml` file, please visit [here](/docs/concepts/how/manifest.md).
- Learn more about `pack` cli from [here](/docs/concepts/how/cli.md).
