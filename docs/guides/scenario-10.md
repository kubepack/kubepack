---
title: Scenarios | Kubepack
menu:
  docs_0.1.0-alpha.0:
    identifier: s10-guides
    name: Schenario 10
    parent: guides
    weight: 70
menu_name: docs_0.1.0-alpha.0
section_menu_id: guides
---

> New to Kubepack? Please start [here](/docs/concepts/README.md).

# Scenario-10


**This docs explain how Pack works inside cluster.**
***

In this scenario, we'll do following things.

1. Create a git repository.
   - In this repository, require [test-kubed](https://github.com/kubepack/test-kubed) through manifest.yaml file.
   - Run `$ pack dep` to get the dependencies and `$ pack edit -s <filepath>` to make desired changes.
   - Then, run `$ pack up` to final version under `_outlook` folder.
   - Last, commit our changes to git repository.

2.  Now, I write a pod yaml, you can see it [here](https://raw.githubusercontent.com/kubepack/kubepack/test-mount/_testdata/test-10/pod.yaml).
In this pod, our above git repository mounted as volume path and
 image [a8uhnf/git-mount:1.0.0](https://cloud.docker.com/swarm/a8uhnf/repository/docker/a8uhnf/git-mount/tags),
is checking the mounted path if their is `_outlook` folder then it'll apply `$ kubectl apply -R -f <_outlook folder path>`.


## Step by Step Guide

First, create a git repository

Create a `manifest.yaml` file in your git repository. Your manifest.yaml file will look like below.

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

It depends on [test-kubed](github.com/kubepack/test-kubed)'s master branch.

Now, run `$ pack dep`. This command will get all the dependencies and place under `_vendor` folder.

```console
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

Now, you have all the dependencies.

Now, suppose you want to edit `deployment.yaml` file and make the replicas from 1 to 2.


Below command will open the `deployment.yaml` file in editor. Then made the changes.
```console
    $ pack edit -s _vendor/github.com/kubepack/test-kubed/deployment.yaml
```

This command will generate a patch file under `patch` folder.

```console
    $ tree patch/

    patch/
    └── github.com
        └── kubepack
            └── test-kubed
                └── deployment.yaml

    3 directories, 1 file
```


Then, run `$ pack up`, which will combine original and patch file and place under `_outlook` folder.

```console
    $ tree _outlook/

    _outlook/
    └── github.com
        └── kubepack
            └── test-kubed
                ├── deployment.yaml
                ├── kubed-config.yaml
                └── service.yaml

    3 directories, 3 files
```

Now, last step, commit the whole thing and push it git repository.



Now, see below [this](https://raw.githubusercontent.com/kubepack/kubepack/test-mount/_testdata/test-10/pod.yaml) yaml file.

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

change the above yaml file's `gitRepo.Repository` and `gitRepo.revision` to your repository location and revision.

```console
    $ kc apply -f https://raw.githubusercontent.com/kubepack/kubepack/test-mount/_testdata/test-10/pod.yaml
    pod "server" created
```

This pod mount your git repository in /mypath in the container and if their is exist any `_outlook` folder, then it'll `$ kubeclt apply -R -f <outlook filepath>`.
You can check actual implementation [here](https://github.com/a8uhnf/git-mount/blob/master/main.go).

Now, you can see the all the desired kubernetes object in your cluster.


## Next Steps

- Want to publish apps using Kubepack? Please visit [here](/docs/concepts/how/publisher.md).
- Want to consume apps published using Kubepack? Please visit [here](/docs/concepts/how/user.md).
- To learn about `manifest.yaml` file, please visit [here](/docs/concepts/how/manifest.md).
- Learn more about `pack` cli from [here](/docs/concepts/how/cli.md).
