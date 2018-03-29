---
title: Using Kubepack CLI
menu:
  docs_0.1.0-alpha.2:
    identifier: cli-how
    name: CLI
    parent: how-concepts
    weight: 25
menu_name: docs_0.1.0-alpha.2
section_menu_id: concepts
---

> New to Kubepack? Please start [here](/docs/concepts/README.md).

# How up command works

`pack up -f .` command combine `manifests/vendor` and `manifests/patch` folder and generate `manifests/output` folder. In this command, it generates a DAG(Directed Acyclic Graph) 
from `dependency-list.yaml` file. 

It describes how dependency chain exists. 

It works in couples of step.

 - Combine `manifests/vendor` and `manifests/patch` to `manifests/output` folder.
 - Generates a DAG(Directed Acyclic Graph), discovers who depend on who, From `dependency-list.yaml`.
 - With generation of DAG, also generate a install.sh file, contains commands to deploy final `manifests/output` folder.
 There could be separate deploy script for some repository, then that script will run instead of default `kubectl apply -R -f .`
 
 
## [Example]()  