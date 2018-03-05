---
title: Pack
menu:
  docs_0.1.0-alpha.1:
    identifier: pack
    name: Pack
    parent: reference
    weight: 0

menu_name: docs_0.1.0-alpha.1
section_menu_id: reference
aliases:
  - /docs/0.1.0-alpha.1/reference/

---
## pack

Secure Lightweight Kubernetes Package Manager

### Synopsis

Secure Lightweight Kubernetes Package Manager

### Options

```
      --alsologtostderr                  log to standard error as well as files
      --analytics                        Send analytical events to Google Guard (default true)
      --as string                        Username to impersonate for the operation
      --as-group stringArray             Group to impersonate for the operation, this flag can be repeated to specify multiple groups.
      --certificate-authority string     Path to a cert file for the certificate authority
      --client-certificate string        Path to a client certificate file for TLS
      --client-key string                Path to a client key file for TLS
      --cluster string                   The name of the kubeconfig cluster to use
      --context string                   The name of the kubeconfig context to use
  -h, --help                             help for pack
      --insecure-skip-tls-verify         If true, the server's certificate will not be checked for validity. This will make your HTTPS connections insecure
      --kube-version string              name of the kubeconfig context to use
      --kubeconfig string                Path to the kubeconfig file to use for CLI requests.
      --log-backtrace-at traceLocation   when logging hits line file:N, emit a stack trace (default :0)
      --log-dir string                   If non-empty, write log files in this directory
      --logtostderr                      log to standard error instead of files
  -n, --namespace string                 If present, the namespace scope for this CLI request
      --password string                  Password for basic authentication to the API server
      --request-timeout string           The length of time to wait before giving up on a single server request. Non-zero values should contain a corresponding time unit (e.g. 1s, 2m, 3h). A value of zero means don't timeout requests. (default "0")
  -s, --server string                    The address and port of the Kubernetes API server
      --stderrthreshold severity         logs at or above this threshold go to stderr (default 2)
      --token string                     Bearer token for authentication to the API server
      --user string                      The name of the kubeconfig user to use
      --username string                  Username for basic authentication to the API server
  -v, --v Level                          log level for V logs
      --vmodule moduleSpec               comma-separated list of pattern=N settings for file-filtered logging
```

### SEE ALSO

* [pack base64](/docs/reference/pack_base64.md)	 - Base64 encode/decode input text
* [pack dep](/docs/reference/pack_dep.md)	 - Pulls dependent app manifests
* [pack edit](/docs/reference/pack_edit.md)	 - Edit resource definition
* [pack envsubst](/docs/reference/pack_envsubst.md)	 - Emulates bash environment variable substitution for input text
* [pack has-keys](/docs/reference/pack_has-keys.md)	 - Checks configmap/secret has a set of given keys
* [pack init](/docs/reference/pack_init.md)	 - Initialize kubepack and create manifest.yaml file
* [pack install](/docs/reference/pack_install.md)	 - Install as kubectl plugin
* [pack jsonpath](/docs/reference/pack_jsonpath.md)	 - Print value of jsonpath for input text
* [pack semver](/docs/reference/pack_semver.md)	 - Print sanitized semver version
* [pack ssl](/docs/reference/pack_ssl.md)	 - Utility commands for SSL certificates
* [pack up](/docs/reference/pack_up.md)	 - Compiles patches and vendored manifests into final resource definitions
* [pack validate](/docs/reference/pack_validate.md)	 - Validate manifests/output folder
* [pack version](/docs/reference/pack_version.md)	 - Prints binary version number.
* [pack wait-until-ready](/docs/reference/pack_wait-until-ready.md)	 - Wait until resource is ready

