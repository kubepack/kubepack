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

`pack up -f .` command combines `manifests/vendor` and `manifests/patch` folders and generates `manifests/output` folder. It works in couples of step.

 - Combines `manifests/vendor` and `manifests/patch` to `manifests/output` folder.
 - Generates a DAG(Directed Acyclic Graph) from `dependency-list.yaml`. From this dependency graph, it generates a `install.sh` file. This installer script contains commands to deploy `manifests/output` folder. Each parent package can provide their own `install.sh` script. If no script is provided, `kubectl apply -R -f .` command will be used to install a package.

- [Example](/docs/_testdata/test-11)

## Next Steps

- Want to publish apps using Kubepack? Please visit [here](/docs/concepts/how/publisher.md).
- Want to consume apps published using Kubepack? Please visit [here](/docs/concepts/how/user.md).
- To learn about `dependency-list.yaml` file, please visit [here](/docs/concepts/how/manifest.md).
- Learn more about `pack` cli from [here](/docs/concepts/how/cli.md).
