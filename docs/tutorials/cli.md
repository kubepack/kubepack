> New to Pack? Please start [here](/docs/tutorials/README.md).

# Use Pack Cli

### How to Get Dependencies

```console
    $ pack dep
```
command will get dependencies defined under `dependencies` field in `manifest.yaml` file. You can get specific version, branch or revision.
See tutorial of [manifest.yaml](/docs/tutorials/manifest.md). All the dependencies will appear in `_vendor` folder.
You can get verbose output with `--v=10` or `-v 10` flag. 

### Edit File from _vendor Folder
```console
    $ pack edit -s <filepath>
```
command edit file, exists in `_vendor` folder and generate patch in `patch` folder. 
This patch file-path will be same as `_vendor` folder. 

**Note: `filepath`: is relative file path.**

### Combine _vendor and patch files

```console
    $ pack up
```
command combine files from `patch` and `_vendor` folder. This combination of `patch` and `_vendor` files appear in `outlook` folder.
