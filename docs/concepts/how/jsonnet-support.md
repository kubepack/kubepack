---
title: How Kubepack support jsonnet
menu:
  docs_0.1.0-alpha.2:
    identifier: jsonnet-support
    name: Jsonnet Support
    parent: how-concepts
    weight: 30
menu_name: docs_0.1.0-alpha.2
section_menu_id: concepts
---

> New to Kubepack? Please start [here](/docs/concepts/README.md).

# Pack JSONNET Support

Pack support jsonnet. 
Pack [publisher](/docs/concepts/how/publisher.md) can write kubernetes resource's definition in jsonnet format.

Then dependant users can require that repository via [dependency-list.yaml](/docs/concepts/how/manifest.md) and simply `$ pack dep` command. 

This will bring all the dependencies in `manifests/vendor` folder. Remainder, publisher's repository may contains jsonnet file,
 but it'll appear in kubernetes resource's yaml format under user's `manifests/vendor` folder.
 
## Learn More About jsonnet Support

- [Here](/docs/guides/scenario-9.md) How jsonnet works in **Pack** and appears in `manifests/vendor` folder.
- [Here](/docs/guides/scenario-10.md) How jsonnet works in **Pack**, in more complex scenario than previous one.


## Next Steps

- Want to publish apps using Kubepack? Please visit [here](/docs/concepts/how/publisher.md).
- Want to consume apps published using Kubepack? Please visit [here](/docs/concepts/how/user.md).
- To learn about `dependency-list.yaml` file, please visit [here](/docs/concepts/how/manifest.md).
- Learn more about `pack` cli from [here](/docs/concepts/how/cli.md).
- Learn more about **Pack** validation [here](/docs/concepts/how/validation.md).
  
