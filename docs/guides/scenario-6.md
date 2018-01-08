---
title: Scenarios | Kubepack
menu:
  docs_0.1.0-alpha.0:
    identifier: s6-guides
    name: Scenario 6
    parent: guides
    weight: 70
menu_name: docs_0.1.0-alpha.0
section_menu_id: guides
---

> New to Kubepack? Please start [here](/docs/concepts/README.md).

# Scenario-6

**This docs explain how Pack's fork works.**
***

This section explain [test-6](https://github.com/kubepack/kubepack/tree/master/_testdata/test-6).

If you look into this test's `manifest.yaml` file.

```console
$ cat manifest.yaml

package: github.com/kubepack/kubepack/_testdata/test-6
owners:
- name: Appscode
  email: team@appscode.com
dependencies:
- package: github.com/kubepack/kube-a
  branch: test-8
  fork: https://github.com/kubepack/kube-d.git
- package: github.com/kubepack/kube-b
  branch: test-8
  fork: github.com/kubepack/kube-c
```

See image below, which describe whole dependency.
![alt text](/_testdata/test-6/test-6.jpg)

Explanation of image:

1. `kube-c` and `kube-d` both has patch of both `kube-a` and `kube-b`.
2. This test depends on two repository.
  - `kube-a` from fork `kube-d`. Means `kube-a` is which exist in `_vendor` folder in `kube-d` repository. Also, applied the patch.
  - `kube-b` from fork `kube-c`. Means `kube-b` is which exist in `_vendor` folder in `kube-c` repository. Also, applied the patch.

Now, `$ pack dep` command get the dependencies and place under `_vendor` folder.
Here, `kube-a` from fork `kube-d` and `kube-b` from fork `kube-c`.


## Next Steps

- Want to publish apps using Kubepack? Please visit [here](/docs/concepts/how/publisher.md).
- Want to consume apps published using Kubepack? Please visit [here](/docs/concepts/how/user.md).
- To learn about `manifest.yaml` file, please visit [here](/docs/concepts/how/manifest.md).
- Learn more about `pack` cli from [here](/docs/concepts/how/cli.md).
