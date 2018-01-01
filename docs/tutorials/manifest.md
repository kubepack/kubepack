# Pack

### manifest.yaml

`manifest.yaml` is metadata file for pack like [dep](https://github.com/golang/dep)'s Gopkg.toml and [glide](https://github.com/Mastermind/glide)'s glide.yaml.
 
 The `manifest.yaml` contains below element:
 
 1. It names the current package.
 2. It declares the external dependencies
 
 A brief `manifest.yaml` file looks like this:
 
 ```
package: github.com/kubepack/pack
owners:
- name: AppsCode
  email: team@appscode.com
dependencies:
- package: github.com/kubepack/kube-a
- package: github.com/kubepack/kube-b
  version: ^1.2.0
  repo:    git@github.com:kubepack/kube-c
- package: github.com/codegangsta/cli
  version: f89effe81c1ece9c5b0fda359ebd9cf65f169a51
- package: github.com/Masterminds/semver
  version: ^1.0.0
testImport:
- package: github.com/arschles/assert 
```
 
  - package: The top level package is the location in the `GOPATH`. 
  This is used for things such as making sure an import isn't also importing the top level package.
  - owners: The owners is a list of one or more owners for the project. This can be a person or organization and is useful for things like notifying the owners of a security issue without filing a public bug.
  - dependencies: A list of external package needs to import. Each package can include:
    - package: 
    