---
title: Scenarios | Kubepack
menu:
  docs_0.1.0-alpha.1:
    identifier: s1-guides
    name: Scenario 1
    parent: guides
    weight: 70
menu_name: docs_0.1.0-alpha.1
section_menu_id: guides
---

> New to Kubepack? Please start [here](/docs/concepts/README.md).

# Scenario-1

**This docs trying to explain the behavior of Pack**
***

This section explain [test-1](https://github.com/kubepack/pack/tree/master/docs/_testdata/test-1).

If you look into this test's `manifest.yaml` file.

```console
$ cat manifest.yaml

package: github.com/kubepack/pack/docs/_testdata/test-1
owners:
- name: Appscode
  email: team@appscode.com
dependencies:
- package: github.com/kubepack/kube-a
  branch: test-1

```

See image below, which describe whole dependency.

![alt text](/docs/_testdata/test-1/test-1.jpg)

Explanation of image:

1. This test directly depends on branch `test-1` of `kube-a` repository.

2. `kube-a`'s depends on branch `test-1` of `kube-b`.
See this manifest.yaml file [here](https://github.com/kubepack/kube-a/blob/test-1/manifest.yaml)

3. `kube-b`'s depends on branch `test-1` of `kube-c`.
See this manifest.yaml file [here](https://github.com/kubepack/kube-b/blob/test-1/manifest.yaml)

4. `kube-c` has no dependency.
See this manifest.yaml file [here](https://github.com/kubepack/kube-c/blob/test-1/manifest.yaml)


Now, `$ kubectl plugin pack dep` command will get all the dependencies and place it under `manifests/vendor` folder.

## Next Steps

- Want to publish apps using Kubepack? Please visit [here](/docs/concepts/how/publisher.md).
- Want to consume apps published using Kubepack? Please visit [here](/docs/concepts/how/user.md).
- To learn about `manifest.yaml` file, please visit [here](/docs/concepts/how/manifest.md).
- Learn more about `pack` cli from [here](/docs/concepts/how/cli.md).
