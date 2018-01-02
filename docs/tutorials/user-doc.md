# Users Docs

## How To Use

If you want to use others application with pack, then follow below instruction:

1. Create a git repository.
2. Create manifest.yaml file in the repository. See [manifest.yaml](/docs/tutorials/manifest.md) doc
3. Add all the dependencies under `dependencies` in manifest.yaml file.
4. Run `pack dep -v 10` to get all the dependencies in `_vendor` folder.
5. Run `pack edit -s <filepath>`, if you want to change some file from `_vendor` folder. 
This command will generate a patch under `patch` folder. `_vendor` will be unchanged.
6. Run `pack up` to combine `patch` and `_vendor` folder files. 
And final combination will be under `_outlook` folder.
7. Now, all is need to do `kubectl apply -R -f ./_outlook/`.
 Then, your desired application will be deployed in kubernetes cluster.   

**Note: If you're application developer and needs to expose your application using pack, follow [this](/docs/tutorials/dev-doc.md) doc**