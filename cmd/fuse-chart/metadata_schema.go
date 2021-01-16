package main

const valuesMetadataSchema = `properties:
  metadata:
    properties:
      release:
        properties:
          name:
            type: string
          namespace:
            type: string
        required:
        - name
        - namespace
        type: object
      resource:
        description: ResourceID identifies a resource
        properties:
          group:
            type: string
          kind:
            description: Kind is the serialized kind of the resource.  It is normally
              CamelCase and singular.
            type: string
          name:
            description: 'Name is the plural name of the resource to serve.  It must
              match the name of the CustomResourceDefinition-registration too: plural.group
              and it must be all lowercase.'
            type: string
          scope:
            description: ResourceScope is an enum defining the different scopes available
              to a custom resource
            type: string
          version:
            type: string
        required:
        - group
        - kind
        - name
        - scope
        - version
        type: object
    required:
    - release
    - resource
    type: object
type: object`
