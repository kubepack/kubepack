> New to Pack? Please start [here](/docs/tutorials/README.md).

# Scenario-8


**This docs explain how Pack's fork works.**
***

This section explain [test-8](https://github.com/kubepack/pack/tree/master/_testdata/test-8).

If you look into this test's `manifest.yaml` file.

```console
$ cat manifest.yaml

package: github.com/kubepack/pack/_testdata/test-7
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
![alt text](/_testdata/test-8/test-8.jpg)

Explanation of image:

1. `kube-c` and `kube-d` both has patch of both `kube-a` and `kube-b`. 
2. This test depends on two repository.
  - `kube-a` from fork `kube-d`. Means `kube-a` is which exist in `_vendor` folder in `kube-d` repository. Also, applied the patch.
  - `kube-b` from fork `kube-c`. Means `kube-b` is which exist in `_vendor` folder in `kube-c` repository. Also, applied the patch.

Now, `$ pack dep` command get the dependencies and place under `_vendor` folder.
Here, `kube-a` from fork `kube-d` and `kube-b` from fork `kube-c`.


# Next Steps

- Learn about `manifest.yaml` file. Please visit [here](/docs/tutorials/manifest.md).
- Learn about `pack` cli. Please visit [here](/docs/tutorials/cli.md)
