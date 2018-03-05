---
title: Scenarios | Kubepack
menu:
  docs_0.1.0-alpha.0:
    identifier: s3-guides
    name: Scenario 3
    parent: guides
    weight: 70
menu_name: docs_0.1.0-alpha.0
section_menu_id: guides
---

> New to Kubepack? Please start [here](/docs/concepts/README.md).

# Scenario-3

**This docs trying to explain the behavior of Pack**
***

This section explain [test-3](https://github.com/kubepack/pack/tree/master/docs/_testdata/test-3).

If you look into this test's `manifest.yaml` file.

```console
$ cat manifest.yaml

package: github.com/kubepack/pack/docs/_testdata/test-3
owners:
- name: Appscode
  email: team@appscode.com
dependencies:
- package: github.com/kubepack/kube-a
  branch: test-3

```

See image below, which describe whole dependency.

![alt text](/docs/_testdata/test-3/test-3.jpg)


Explanation of image:

1. This test directly depends on branch `test-3` of [kube-a](https://github.com/kubepack/kube-a/tree/test-3) repository.
2. `kube-a` depends on branch `test-3` of [kube-b](https://github.com/kubepack/kube-b/tree/test-3) repository.
See this manifest.yaml file [here](https://github.com/kubepack/kube-a/blob/test-3/manifest.yaml).
3. `kube-b` depends on branch `test-3` of `kube-c` repository.
See this manifest.yaml file [here](https://github.com/kubepack/kube-b/blob/test-3/manifest.yaml).
4. `kube-c`'s has no dependency.
See this manifest.yaml file [here](https://github.com/kubepack/kube-c/blob/test-3/manifest.yaml).

Here, both `kube-a` and `kube-b` has patch of repository `kube-c`'s [nginx-deployment.yaml file](https://github.com/kubepack/kube-c/blob/test-3/nginx-deployment.yaml). You can find these patches here:

- [kube-a](https://github.com/kubepack/kube-a/blob/test-3/patch/github.com/kubepack/kube-c/nginx-deployment.yaml)
- [kube-b](https://github.com/kubepack/kube-b/blob/test-3/patch/github.com/kubepack/kube-c/nginx-deployment.yaml)

Now, run `$ pack dep` command. This will vendor branch `test-3` of all the dependencies `kube-a`, `kube-b` and `kube-c`.
As, `kube-a` and `kube-b` both contain patch of `kube-c`'s [nginx-deployment.yaml file](https://github.com/kubepack/kube-c/blob/test-3/nginx-deployment.yaml).
This file is the combination of both patch and original file.

## Next Steps

- Want to publish apps using Kubepack? Please visit [here](/docs/concepts/how/publisher.md).
- Want to consume apps published using Kubepack? Please visit [here](/docs/concepts/how/user.md).
- To learn about `manifest.yaml` file, please visit [here](/docs/concepts/how/manifest.md).
- Learn more about `pack` cli from [here](/docs/concepts/how/cli.md).
