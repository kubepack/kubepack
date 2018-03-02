---
title: Scenarios | Kubepack
menu:
  docs_0.1.0-alpha.0:
    identifier: s8-guides
    name: Scenario 8
    parent: guides
    weight: 75
menu_name: docs_0.1.0-alpha.0
section_menu_id: guides
---

> New to Kubepack? Please start [here](/docs/concepts/README.md).

# Scenario-8


**This docs explain how Pack works inside cluster.**
***

In this scenario, we'll do following things.

1. Create a git repository.
   - This repository requires [test-kubed](https://github.com/kubepack/test-kubed) through `manifest.yaml` file.
   - Run `$ pack dep` to get the dependencies and `$ pack edit -s <filepath>` to make desired changes.
   - Then, run `$ pack up` to final version under `manifests/output` folder.
   - Last, commit our changes to git repository.

2.  Now, I write a pod yaml, you can see it [here](https://raw.githubusercontent.com/kubepack/kubepack/master/docs/_testdata/test-8/pod.yaml).
In this pod, our above git repository mounted as volume path. Image [a8uhnf/git-mount:1.0.0](https://hub.docker.com/r/a8uhnf/git-mount/tags/) checks the mounted path. If there is an `manifest/output` folder then it'll apply `$ kubectl apply -R -f <manifest/output folder path>`.


## Step by Step Guide

First, create a git repository

Create a `manifest.yaml` file in your git repository. Your `manifest.yaml` file will look like below.

```console
    $ cat manifest.yaml

package: YOUR_REPO_LOCATION # github.com/packsh/tasty-kube
owners:
- name: # Appscode
  email: # team@appscode.com
dependencies:
 - package: github.com/kubepack/test-kubed
   branch: master
```

It depends on [test-kubed](https://github.com/kubepack/test-kubed)'s master branch.

Now, run `$ pack dep`. This command will get all the dependencies and place under `manifests/vendor` folder.

```console
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

Now, you have all the dependencies.

Now, suppose you want to edit `deployment.yaml` file and make the replicas from 1 to 2.


Below command will open the `deployment.yaml` file in editor. Then made the changes.
```console
    $ kubepack edit -s manifests/vendor/github.com/kubepack/test-kubed/manifests/app/deployment.yaml
```

This command will generate a patch file under `patch` folder.

```console
    $ tree manifests/patch/

    manifests/patch/
    └── github.com
        └── kubepack
            └── test-kubed
                └── manifests
                    └── app
                        └── deployment.yaml
    
    5 directories, 1 file
```


Then, run `$ pack up`, which will combine original and patch file and place under `manifests/output` folder.

```console
    $ tree manifests/output/

    manifests/output/
    └── github.com
        └── kubepack
            └── test-kubed
                ├── deployment.yaml
                ├── kubed-config.yaml
                └── service.yaml

    3 directories, 3 files
```

Now, last step, commit the whole thing and push it git repository.


Now, see below [this](https://raw.githubusercontent.com/kubepack/kubepack/master/docs/_testdata/test-8/pod.yaml) yaml file.

```console
    apiVersion: v1
    kind: Pod
    metadata:
      name: server
    spec:
      containers:
      - image: a8uhnf/git-mount:1.0.0
        imagePullPolicy: Always
        name: git-mount
        resources: {}
        volumeMounts:
        - mountPath: /mypath
          name: git-volume
      volumes:
      - gitRepo:
          repository: YOUR_GIT_REPO_LOCATION # https://github.com/kubepack/kube-a.git
          revision: GIT_REPO_REVISION_NUMBER # c90e98d6c0a6143c19a6e3a575befbdfa170fa00
        name: git-volume
    status: {}
```

change the above yaml file's `gitRepo.repository` and `gitRepo.revision` to your repository location and revision.

```console
    $ kubectl apply -f https://raw.githubusercontent.com/kubepack/kubepack/master/docs/_testdata/test-8/pod.yaml
    pod "server" created
```

This pod mount your git repository in /mypath in the container and if their is exist any `manifests/output` folder, then it'll `$ kubectl apply -R -f <output filepath>`.
You can check actual implementation [here](https://github.com/kubepack/git-mount/blob/master/main.go).

Now, you can see the all the desired kubernetes object in your cluster.


## Next Steps

- Want to publish apps using Kubepack? Please visit [here](/docs/concepts/how/publisher.md).
- Want to consume apps published using Kubepack? Please visit [here](/docs/concepts/how/user.md).
- To learn about `manifest.yaml` file, please visit [here](/docs/concepts/how/manifest.md).
- Learn more about `pack` cli from [here](/docs/concepts/how/cli.md).
