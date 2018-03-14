---
title: Pack Ssl Get
menu:
  docs_0.1.0-alpha.2:
    identifier: pack-ssl-get
    name: Pack Ssl Get
    parent: reference
menu_name: docs_0.1.0-alpha.2
section_menu_id: reference
---
## pack ssl get

Get stuff

### Synopsis

Get stuff

### Options

```
  -h, --help   help for get
```

### Options inherited from parent commands

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
  -f, --file string                      filepath
      --insecure-skip-tls-verify         If true, the server's certificate will not be checked for validity. This will make your HTTPS connections insecure
      --kube-version string              name of the kubeconfig context to use
      --kubeconfig string                Path to the kubeconfig file to use for CLI requests.
      --log-backtrace-at traceLocation   when logging hits line file:N, emit a stack trace (default :0)
      --log-dir string                   If non-empty, write log files in this directory
      --logtostderr                      log to standard error instead of files
  -n, --namespace string                 If present, the namespace scope for this CLI request
      --password string                  Password for basic authentication to the API server
  -p, --patch string                     File want to edit
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

* [pack ssl](/docs/reference/pack_ssl.md)	 - Utility commands for SSL certificates
* [pack ssl get ca-cert](/docs/reference/pack_ssl_get_ca-cert.md)	 - Prints self-sgned CA certificate from PEM encoded RSA private key
* [pack ssl get kube-ca](/docs/reference/pack_ssl_get_kube-ca.md)	 - Prints CA certificate for Kubernetes cluster from Kubeconfig

