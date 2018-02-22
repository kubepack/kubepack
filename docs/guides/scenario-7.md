---
title: Scenarios | Kubepack
menu:
  docs_0.1.0-alpha.0:
    identifier: s7-guides
    name: Scenario 7
    parent: guides
    weight: 70
menu_name: docs_0.1.0-alpha.0
section_menu_id: guides
---

> New to Kubepack? Please start [here](/docs/concepts/README.md).

# Scenario-7

**This docs explain how can deploy kubed using Pack.**
***

In this example, you'll see how to deploy [AppsCode kubed](https://github.com/appscode/kubed)
 in minikube using `Pack`.

In this example, we're using [this](https://github.com/kubepack/kubepack/tree/master/docs/_testdata/test-7) test-case.

Below command show the `manifest.yaml` file.
```console
$ cat manifest.yaml
package: github.com/kubepack/kubepack/docs/_testdata/test-7
owners:
- name: Appscode
  email: team@appscode.com
dependencies:
- package: github.com/kubepack/test-kubed
  branch: master
```
`manifest.yaml` file contain [test-kubed](https://github.com/kubepack/test-kubed) as a dependency. `test-kubed` contains
 all the necessary yaml file needs to deploy kubed in minikube cluster.

Now, `$ pack dep` command will pull all the dependencies and place it in `_vendor` folder. If `test-kubed` repository also depend on some other repository then `pack` will get those too.

  ```console
  $ pack dep
  $ tree _vendor/
  _vendor/
  └── github.com
      └── kubepack
          └── test-kubed
              ├── deployment.yaml
              ├── kubed-config.yaml
              ├── manifest.yaml
              └── service.yaml

  3 directories, 4 files
  ```
  Now, all the dependencies in place. Now, we can edit `_vendor` and this will generate patch.

  We're want to change `kubed-config.yaml` file, which is a secret yaml file.

  ```console
    $ cat _vendor/github.com/kubepack/test-kubed/kubed-config.yaml
    apiVersion: v1
    data:
      config.yaml: YXBpU2VydmVyOgogIGFkZHJlc3M6IDo4MDgwCiAgZW5hYmxlUmV2ZXJzZUluZGV4OiB0cnVlCiAgZW5hYmxlU2VhcmNoSW5kZXg6IHRydWUKY2x1c3Rlck5hbWU6IHVuaWNvcm4KZW5hYmxlQ29uZmlnU3luY2VyOiB0cnVlCmV2ZW50Rm9yd2FyZGVyOgogIGNzckV2ZW50czoKICAgIGhhbmRsZTogZmFsc2UKICBpbmdyZXNzQWRkZWQ6CiAgICBoYW5kbGU6IHRydWUKICBub2RlQWRkZWQ6CiAgICBoYW5kbGU6IHRydWUKICByZWNlaXZlcnM6CiAgLSBub3RpZmllcjogTWFpbGd1bgogICAgdG86CiAgICAtIG9wc0BleGFtcGxlLmNvbQogIHN0b3JhZ2VBZGRlZDoKICAgIGhhbmRsZTogdHJ1ZQogIHdhcm5pbmdFdmVudHM6CiAgICBoYW5kbGU6IHRydWUKICAgIG5hbWVzcGFjZXM6CiAgICAtIGt1YmUtc3lzdGVtCmphbml0b3JzOgotIGVsYXN0aWNzZWFyY2g6CiAgICBlbmRwb2ludDogaHR0cHM6Ly9lbGFzdGljc2VhcmNoLWxvZ2dpbmcua3ViZS1zeXN0ZW06OTIwMAogICAgbG9nSW5kZXhQcmVmaXg6IGxvZ3N0YXNoLQogICAgc2VjcmV0TmFtZTogZWxhc3RpY3NlYXJjaC1sb2dnaW5nLWNlcnQKICBraW5kOiBFbGFzdGljc2VhcmNoCiAgdHRsOiAyMTYwaDBtMHMKLSBpbmZsdXhkYjoKICAgIGVuZHBvaW50OiBodHRwczovL21vbml0b3JpbmctaW5mbHV4ZGIua3ViZS1zeXN0ZW06ODA4NgogIGtpbmQ6IEluZmx1eERCCiAgdHRsOiAyMTYwaDBtMHMKbm90aWZpZXJTZWNyZXROYW1lOiBub3RpZmllci1jb25maWcKcmVjeWNsZUJpbjoKICBoYW5kbGVVcGRhdGVzOiBmYWxzZQogIHBhdGg6IC90bXAva3ViZWQvdHJhc2gKICByZWNlaXZlcnM6CiAgLSBub3RpZmllcjogTWFpbGd1bgogICAgdG86CiAgICAtIG9wc0BleGFtcGxlLmNvbQogIHR0bDogMTY4aDBtMHMKc25hcHNob3R0ZXI6CiAgZ2NzOgogICAgYnVja2V0OiByZXN0aWMKICAgIHByZWZpeDogbWluaWt1YmUKICBzYW5pdGl6ZTogdHJ1ZQogIHNjaGVkdWxlOiAnQGV2ZXJ5IDZoJwogIHN0b3JhZ2VTZWNyZXROYW1lOiBzbmFwLXNlY3JldAo=
    kind: Secret
    metadata:
      creationTimestamp: null
      labels:
        app: kubed
      name: kubed-config
      namespace: kube-system
```

We'll change `config.yaml` under `data` field. `config.yaml` value will be `YXBpU2VydmVyOgogIGFkZHJlc3M6IDo4MDgwCiAgZW5hYmxlUmV2ZXJzZUluZGV4OiB0cnVlCiAgZW5hYmxlU2VhcmNoSW5kZXg6IHRydWUKY2x1c3Rlck5hbWU6IHVuaWNvcm4KZW5hYmxlQ29uZmlnU3luY2VyOiB0cnVlCg==`

```console
$ pack edit -s _vendor/github.com/kubepack/test-kubed/kubed-config.yaml
```

Above command will open file in editor.
 Then, change `config.yaml` to above value. This will generate a patch in `patch` folder.

 Below `$ pack up` command will combine `patch` and `_vendor` folder files and dump in `_outlook` folder.

 ```console
 $ pack up
 $ kubectl apply -R -f _outlook/
 ```
 `$ kubectl apply -R -f _outlook/` command will deploy kubed in minikube cluster.



## Next Steps

- Want to publish apps using Kubepack? Please visit [here](/docs/concepts/how/publisher.md).
- Want to consume apps published using Kubepack? Please visit [here](/docs/concepts/how/user.md).
- To learn about `manifest.yaml` file, please visit [here](/docs/concepts/how/manifest.md).
- Learn more about `pack` cli from [here](/docs/concepts/how/cli.md).
