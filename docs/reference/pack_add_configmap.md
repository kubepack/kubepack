---
title: Pack Add Configmap
menu:
  docs_0.1.0-alpha.2:
    identifier: pack-add-configmap
    name: Pack Add Configmap
    parent: reference
menu_name: docs_0.1.0-alpha.2
section_menu_id: reference
---
## pack add configmap

Adds a configmap to the manifest.

### Synopsis

Adds a configmap to the manifest.

```
pack add configmap NAME [--from-file=[key=]source] [--from-literal=key1=value1] [flags]
```

### Examples

```

	# Adds a configmap to the Manifest (with a specified key)
	kinflate add configmap my-configmap --from-file=my-key=file/path --from-literal=my-literal=12345

	# Adds a configmap to the Manifest (key is the filename)
	kinflate add configmap my-configmap --from-file=file/path

	# Adds a configmap from env-file
	kinflate add configmap my-configmap --from-env-file=env/path.env

```

### Options

```
      --from-env-file string       Specify the path to a file to read lines of key=val pairs to create a configmap (i.e. a Docker .env file).
      --from-file stringSlice      Key file can be specified using its file path, in which case file basename will be used as configmap key, or optionally with a key and file path, in which case the given key will be used.  Specifying a directory will iterate each named file in the directory whose basename is a valid configmap key.
      --from-literal stringArray   Specify a key and literal value to insert in configmap (i.e. mykey=somevalue)
  -h, --help                       help for configmap
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

* [pack add](/docs/reference/pack_add.md)	 - Adds configmap/resource/secret to the manifest.

