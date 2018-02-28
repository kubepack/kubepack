---
title: Scenarios | Kubepack
menu:
  docs_0.1.0-alpha.0:
    identifier: s5-guides
    name: Scenario 5
    parent: guides
    weight: 70
menu_name: docs_0.1.0-alpha.0
section_menu_id: guides
---

> New to Kubepack? Please start [here](/docs/concepts/README.md).

# Scenario-5

**This docs trying to explain the behavior of Pack**
***

This section explain [test-5](https://github.com/kubepack/kubepack/tree/master/docs/_testdata/test-5).

If you look into this test's `manifest.yaml` file.

```console
$ cat manifest.yaml

package: github.com/kubepack/kubepack/docs/_testdata/test-5
owners:
- name: Appscode
  email: team@appscode.com
dependencies:
- package: github.com/kubepack/kube-a
  branch: test-5
- package: github.com/kubepack/kube-b
  branch: test-5
```

Here, [test-5](https://github.com/kubepack/kubepack/tree/master/docs/_testdata/test-5) depends on two repositories.

1. branch `test-5` of [kube-a](https://github.com/kubepack/kube-a/tree/test-5).
2. branch `test-5` of [kube-b](https://github.com/kubepack/kube-b/tree/test-5).

Both of the above repository contains the patch of repository [kube-c](https://github.com/kubepack/kube-c/tree/test-5)'s
 branch `test-5` in same file (nginx-deployment.yaml).

 See the image.
 
![alt text](/docs/_testdata/test-5/test-5.jpg)

You can see the both patch below

```console
# kube-a contains this patch of kube-c

spec:
  replicas: 2
```

```console
# kube-b contains this patch of kube-c

apiVersion: apps/v1beta2
```

When run `pack dep` command, following things happen.

1. Get all the dependencies, reading `manifest.yaml` file.
2. As, `kube-a` and `kube-b` both contains patch of repository `kube-c`,
`kube-c` in `manifests/vendor` folder is combination of both patches and original file.


## Next Steps

- Want to publish apps using Kubepack? Please visit [here](/docs/concepts/how/publisher.md).
- Want to consume apps published using Kubepack? Please visit [here](/docs/concepts/how/user.md).
- To learn about `manifest.yaml` file, please visit [here](/docs/concepts/how/manifest.md).
- Learn more about `pack` cli from [here](/docs/concepts/how/cli.md).
