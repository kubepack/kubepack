> New to Pack? Please start [here](/docs/tutorials/README.md).

# Scenario-3

**This docs trying to explain the behavior of Pack**
***

This section explain [test-3](https://github.com/kubepack/pack/tree/master/_testdata/test-3).

If you look into this test's `manifest.yaml` file.

```console
$ cat manifest.yaml

package: github.com/kubepack/pack/_testdata/test-3
owners:
- name: Appscode
  email: team@appscode.com
dependencies:
- package: github.com/kubepack/kube-a
  branch: test-3
- package: github.com/kubepack/kube-b
  branch: test-3
```

Here, [test-3](https://github.com/kubepack/pack/tree/master/_testdata/test-3) depends on two repositories.
1. [kube-a](https://github.com/kubepack/kube-a) of branch `test-3`.
2. [kube-b](https://github.com/kubepack/kube-b) of branch `test-3`.

Both of the above repository(kube-a and kube-b) require dependency of
 [kube-c](https://github.com/kubepack/kube-c) but two different branch.
 `kube-a` depends on `kube-c` of branch `test-1` and `kube-b` depends on `kube-c` of branch `master`. 
 
 To clarify, see image.
 ![alt text](/_testdata/test-3/test-3.jpg)
 
 Now, when run `$ pack dep --v=10` command, `pack` could not resolve dependencies. As, this dependencies contradict with each other.
  Give below error.
  
  ```console
  $ pack dep --v=10
  I0103 15:46:39.430663    5923 logs.go:19] No versions of github.com/kubepack/kube-b met constraints:
          master: Could not introduce github.com/kubepack/kube-b@master, as it is not allowed by constraint test-3 from project github.com/a8uhnf/pack/_testdata/test-3.
          dep-c: Could not introduce github.com/kubepack/kube-b@dep-c, as it is not allowed by constraint test-3 from project github.com/a8uhnf/pack/_testdata/test-3.
          test-1: Could not introduce github.com/kubepack/kube-b@test-1, as it is not allowed by constraint test-3 from project github.com/a8uhnf/pack/_testdata/test-3.
          test-2: Could not introduce github.com/kubepack/kube-b@test-2, as it is not allowed by constraint test-3 from project github.com/a8uhnf/pack/_testdata/test-3.
          test-3: Could not introduce github.com/kubepack/kube-b@test-3, as it has a dependency on github.com/kubepack/kube-c with constraint master, which has no overlap with existing constraint test-1 from github.com/kubepack/kube-a@test-3
          test-6: Could not introduce github.com/kubepack/kube-b@test-6, as it is not allowed by constraint test-3 from project github.com/a8uhnf/pack/_testdata/test-3.
          test-7: Could not introduce github.com/kubepack/kube-b@test-7, as it is not allowed by constraint test-3 from project github.com/a8uhnf/pack/_testdata/test-3.
          test-8: Could not introduce github.com/kubepack/kube-b@test-8, as it is not allowed by constraint test-3 from project github.com/a8uhnf/pack/_testdata/test-3.

```  

# Next Steps

- Learn about `manifest.yaml` file. Please visit [here](/docs/tutorials/manifest.md).
- Learn about `pack` cli. Please visit [here](/docs/tutorials/cli.md)
