---
title: Using Kubepack as an App User
menu:
  docs_0.1.0-alpha.1:
    identifier: user-how
    name: App Users
    parent: how-concepts
    weight: 15
menu_name: docs_0.1.0-alpha.1
section_menu_id: concepts
---

> New to Kubepack? Please start [here](/docs/concepts/README.md).

# Using Kubepack as an App User

## How To Use

If you want to use others application with pack, then follow below instruction:

1. Create a git repository.
2. Create manifest.yaml file in the repository. See [manifest.yaml](/docs/concepts/how/manifest.md) doc.
3. Add all the dependencies under `dependencies` in manifest.yaml file.
4. Run `kubectl plugin pack dep -v 10` to get all the dependencies in `manifests/vendor` folder.
5. Run `kubectl plugin pack edit -s <filepath>`, if you want to change some file from `manifests/vendor` folder.
This command will generate a patch under `manifests/patch` folder. `manifests/vendor` will be unchanged.
6. Run `kubectl plugin pack up` to combine `manifests/patch` and `manifests/vendor` folder files.
And final combination will be under `manifests/output` folder.
7. Now, all is need to do `kubectl apply -R -f manifests/output/`.
 Then, your desired application will be deployed in kubernetes cluster.

### Example

In this example, you'll see how to deploy [AppsCode kubed](https://github.com/appscode/kubed)
 in minikube using `Pack`.

In this example, we're using [this](https://github.com/kubepack/pack/tree/master/docs/_testdata/test-7/manifest.yaml) test-case.

Below command show the `manifest.yaml` file.
```console
$ cat manifest.yaml
package: github.com/kubepack/pack/_testdata/test-7
owners:
- name: Appscode
  email: team@appscode.com
dependencies:
- package: github.com/kubepack/test-kubed
  branch: master
```
`manifest.yaml` file contain [test-kubed](https://github.com/kubepack/test-kubed) as `dependencies`. `test-kubed` contains
 all the necessary yaml file needs to deploy kubed in minikube cluster.

 Now, `kubectl plugin pack dep` command will pull all the dependencies and place it in `manifests/vendor` folder.
  If `test-kubed` repository also depend on some other repository then `pack` will get that too.

  ```console
  $ kubectl plugin pack dep
  $ tree manifests/vendor/
  
  manifests/vendor/
  └── github.com
      └── kubepack
          └── test-kubed
              ├── manifests
              │   └── app
              │       ├── deployment.yaml
              │       ├── kubed-config.yaml
              │       └── service.yaml
              └── manifest.yaml
  
  5 directories, 4 files
  ```
  Now, all the dependencies in place. Now, we can edit `manifests/vendor` and this will generate patch.

  We're want to change `kubed-config.yaml` file, which is a secret yaml file.

  ```console
    $ cat manifests/vendor/github.com/kubepack/test-kubed/manifests/app/kubed-config.yaml
    
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
$ kubectl plugin pack edit -s manifests/vendor/github.com/kubepack/test-kubed/manifests/app/kubed-config.yaml
```
Above command will open file in editor.
 Then, change `config.yaml` to above value. This will generate a patch in `manifests/patch` folder.

 Below `$ kubectl plugin pack up` command will combine `manifests/patch` and `manifests/vendor` folder files and dump in `manifests/output` folder.

 ```console
 $ kubectl plugin pack up
 $ kubectl apply -R -f manifests/output/
 ```
 `$ kubectl apply -R -f manifests/output/` command will deploy kubed in minikube cluster.


## Next Steps

- Take a look at how Kubepack handles various scenarios [here](/docs/guides).
- Want to publish apps using Kubepack? Please visit [here](/docs/concepts/how/publisher.md).
- To learn about `manifest.yaml` file, please visit [here](/docs/concepts/how/manifest.md).
- Learn more about `pack` cli from [here](/docs/concepts/how/cli.md).
