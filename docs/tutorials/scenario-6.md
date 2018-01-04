> New to Pack? Please start [here](/docs/tutorials/README.md).

# Scenario-6

**This docs trying to explain the behavior of Pack**
***

This section explain [test-6](https://github.com/kubepack/pack/tree/master/_testdata/test-6).

If you look into this test's `manifest.yaml` file.

```console
$ cat manifest.yaml

package: github.com/kubepack/pack/_testdata/test-6
owners:
- name: Appscode
  email: team@appscode.com
dependencies:
- package: github.com/kubepack/kube-a
  branch: test-6
```

See image below, which describe whole dependency.
![alt text](/_testdata/test-6/test-6.jpg)

Explanation of image:

1. This test directly depends on `kube-a` of branch `test-6`.
2. `kube-a`'s depends on `kube-b` of branch `test-6`. 
See this manifest.yaml file [here](https://github.com/kubepack/kube-a/blob/test-6/manifest.yaml)
3. `kube-b`'s depends on `kube-c` of branch `test-6`. 
See this manifest.yaml file [here](https://github.com/kubepack/kube-b/blob/test-6/manifest.yaml)
4. `kube-c`'s depends on none. 
See this manifest.yaml file [here](https://github.com/kubepack/kube-c/blob/test-6/manifest.yaml)


Now, `$ pack dep` command will get all the dependencies and place it under `_vendor` folder. 

# Next Steps

- Learn about `manifest.yaml` file. Please visit [here](/docs/tutorials/manifest.md).
- Learn about `pack` cli. Please visit [here](/docs/tutorials/cli.md)
