---
title: How Kubepack Validate
menu:
  docs_0.1.0-alpha.0:
    identifier: kubepack-validation
    name: Kubepack Validation
    parent: how-concepts
    weight: 35
menu_name: docs_0.1.0-alpha.0
section_menu_id: concepts
---
> New to Kubepack? Please start [here](/docs/concepts/README.md).

# Kubepack Validation

Users can validate their `_outlook` folder through `$ pack validate`. 
Pack uses `openapi spec` for validation. 

Users can provider specific kubernetes version through `kube-version` flag.
 In this case, `_outlook` folder will validate with this kubernetes version `openapi spec`.
 
By default, **Pack** uses  latest stable version of kubernetes.  

## Next Steps

- Want to publish apps using Kubepack? Please visit [here](/docs/concepts/how/publisher.md).
- Want to consume apps published using Kubepack? Please visit [here](/docs/concepts/how/user.md).
- To learn about `manifest.yaml` file, please visit [here](/docs/concepts/how/manifest.md).
- Learn more about `pack` cli from [here](/docs/concepts/how/cli.md).
- Learn more about **Pack** jsonnet-support [here](/docs/concepts/how/jsonnet-support.md).
