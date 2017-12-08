## pack edit

Edit resource definition

### Synopsis


Generates patch via edit command

```
pack edit (filename) [flags]
```

### Options

```
  -h, --help          help for edit
  -s, --src string    File want to edit
  -t, --type string   Type of patch; one of [json merge strategic] (default "strategic")
```

### Options inherited from parent commands

```
      --alsologtostderr                  log to standard error as well as files
      --analytics                        Send analytical events to Google Guard (default true)
      --log_backtrace_at traceLocation   when logging hits line file:N, emit a stack trace (default :0)
      --log_dir string                   If non-empty, write log files in this directory
      --logtostderr                      log to standard error instead of files
      --stderrthreshold severity         logs at or above this threshold go to stderr (default 2)
  -v, --v Level                          log level for V logs
      --vmodule moduleSpec               comma-separated list of pattern=N settings for file-filtered logging
```

### SEE ALSO
* [pack](pack.md)	 - Secure Lightweight Kubernetes Package Manager

