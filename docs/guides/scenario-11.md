---
title: Scenarios | Kubepack
menu:
  docs_0.1.0-alpha.2:
    identifier: s11-guides
    name: Scenario 11
    parent: guides
    weight: 75
menu_name: docs_0.1.0-alpha.2
section_menu_id: guides
---

> New to Kubepack? Please start [here](/docs/concepts/README.md).

#Scenario-10

**This doc is trying to explain how pack up command works and how someone can write user-defined install.sh file.**

![alt text](/docs/_testdata/test-11/test-11.jpg)

To observe, how everything works, needs to clone [this repository](https://kubepack/kube-a). And checkout to `test-11` branch.

You can see `dependency-list.yaml` file in project root.
```console
    $ cat dependency-list.yaml 
    items:
      - package: github.com/kubepack/kube-b
        branch: test-11
      - package: github.com/kubepack/kube-c
        branch: test-11
```



When users execute `pack up -f .` command in this project root, following things happen.

 - Combine `manifests/vendor` and `manifests/patch` to `manifests/output` folder.
 - Generates a DAG(Directed Acyclic Graph), discovers who depend on who, From `dependency-list.yaml`.
 - With generation of DAG, also generate a install.sh file, contains commands to deploy final `manifests/output` folder.
  There could be separate deploy script for some repository, then that script will run instead of default `kubectl apply -R -f .`

 
If you see `manifests/output/install.sh`
 ```console
     cat manifests/output/install.sh
     
     
     #!/bin/bash
     
     
     pushd manifests/output/github.com/kubepack/kube-b
     kubectl apply -R -f .
     popd
     			
     
     pushd manifests/output/github.com/kubepack/kube-c
     kubectl apply -R -f .
     popd
     			
     
     pushd manifests/output/github.com/kubepack/kube-d
     kubectl apply -R -f .
     popd
     			
     
     pushd manifests/output/github.com/kubepack/kube-e
     kubectl apply -R -f .
     popd
     			
     
     pushd manifests/output/github.com/kubepack/kube-f
     kubectl apply -R -f .
     popd
     			
     
     pushd manifests/output/github.com/kubepack/kube-a
     kubectl apply -R -f .
     popd
     	
 ```

 - At first there will be `kubectl apply` command for `kube-b` or `kube-c`, whichever comes first in `dependency-list.yaml`.
 As `kube-b` appears first in `dependency-list.yaml`, `kube-b` will be first, `kube-c` second.
 - `kube-b` will be process first. That's why `kube-d` will appear before `kube-e`. 
 After `kube-d`, `kube-e` will appear.
 - Then `kube-f` and at last, root repo.
 
P.S. If any repository's `manifests/app` folder contain their own `install.sh` file, then instead of default `kubectl apply`, `install.sh` script will run.

Users can use their customize commands for deploy, these customize commands should be in `manifests/app/install.sh` file.


## Next Steps

- Want to publish apps using Kubepack? Please visit [here](/docs/concepts/how/publisher.md).
- Want to consume apps published using Kubepack? Please visit [here](/docs/concepts/how/user.md).
- To learn about `dependency-list.yaml` file, please visit [here](/docs/concepts/how/manifest.md).
- Learn more about `pack` cli from [here](/docs/concepts/how/cli.md).