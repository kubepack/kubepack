---
title: Using Kubepack as an App Publisher
menu:
  docs_0.1.0-alpha.0:
    identifier: publisher-how
    name: App Publishers
    parent: how-concepts
    weight: 15
menu_name: docs_0.1.0-alpha.0
section_menu_id: concepts
---

> New to Kubepack? Please start [here](/docs/concepts/README.md).

# Using Kubepack as an App Publisher

This section contains tutorial on how app developer can use [Pack](https://github.com/kubepack/pack) to expose
theirs app and deploy in kubernetes cluster.

Developer creates a git repo which contains all required yamls and manifest.yaml file.

## Before You Begin

Get the overview and what various fields means, read [manifest.yaml](/docs/guides/manifest.md) tutorial.


## Deploy with Pack

If you want your application to usable through pack,

Needs to follow below instruction:

 - First, create a git repository
 - Add all the required yaml to repository
 - Add manifest.yaml file. This is mandatory file to Pack

Now, anyone can use this repository to deploy your application in their cluster.

### Example

Suppose, you're building a application called `A`. It needs a [deployment](https://raw.githubusercontent.com/kubepack/pack/doc-init/docs/guides/deployment.yaml), [service](https://raw.githubusercontent.com/kubepack/pack/doc-init/docs/guides/service.yaml) and [secret](https://raw.githubusercontent.com/kubepack/pack/doc-init/docs/guides/secret.yaml).

Deployment.yaml:
```
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: pack-deploy
  labels:
    app: kubepack
spec:
  replicas: 3
  selector:
    matchLabels:
      app: kubepack
  template:
    metadata:
      labels:
        app: kubepack
    spec:
      containers:
      - name: nginx
        image: nginx:1.7.9
        ports:
        - containerPort: 80
        volumeMounts:
        - mountPath: /tmp/pack
          name: pack
      volumes:
      - name: pack
        secret:
          secretName: pack-test
```

secrets.yaml:

```
apiVersion: v1
kind: Secret
metadata:
  name: pack-secret
  labels:
    app: kubepack
type: Opaque
data:
  username: YWRtaW4=
  password: MWYyZDFlMmU2N2Rm
```

service.yaml

```
apiVersion: v1
kind: Service
metadata:
  labels:
    app: kubepack
  name: pack-service
spec:
  ports:
  - port: 80
  selector:
    app: kubepack
```

You need to create a repository and put all the [deployment](https://raw.githubusercontent.com/kubepack/pack/doc-init/docs/guides/deployment.yaml), [service](https://raw.githubusercontent.com/kubepack/pack/doc-init/docs/guides/service.yaml) and [secret](https://raw.githubusercontent.com/kubepack/pack/doc-init/docs/guides/secret.yaml) yaml in the repository.
Also, need to create manifest.yaml file in the repository.

So that, others can use it through pack cli.

## Next Steps

 - How can user use this repository? Follow [user](/docs/guides/user-doc.md) doc
 - Learn about `manifest.yaml` file. Please visit [here](/docs/guides/manifest.md).
 - Learn about `pack` cli. Please visit [here](/docs/guides/cli.md)
