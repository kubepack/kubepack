## How to update Hub files in kube-ui-server

```bash
export UI_SERVER_NAMESPACE=kubeops
export UI_SERVER_POD=$(kubectl get pods -n $UI_SERVER_NAMESPACE -l app.kubernetes.io/name=kube-ui-server -o jsonpath={.items[0].metadata.name})

kubectl cp hub $UI_SERVER_POD:/tmp -n $UI_SERVER_NAMESPACE
# verify
kubectl exec -it $UI_SERVER_POD -n $UI_SERVER_NAMESPACE -- ls -l /tmp/hub

LOCATION=hub/resourceeditors/kubedb.com/v1alpha2/elasticsearches.yaml
kubectl cp $LOCATION $UI_SERVER_POD:/tmp/$LOCATION -n $UI_SERVER_NAMESPACE
# verify
kubectl exec -it $UI_SERVER_POD -n $UI_SERVER_NAMESPACE -- cat /tmp/$LOCATION

# trigger reload
kubectl exec -it $UI_SERVER_POD -n $UI_SERVER_NAMESPACE -- sh -c "date > /tmp/hub/resourceeditors/trigger"
# verify
kubectl exec -it $UI_SERVER_POD -n $UI_SERVER_NAMESPACE -- cat /tmp/hub/resourceeditors/trigger
```

## How to modify GET & LIST proxy calls

```bash
$ cd ~/go/src/kmodules.xyz/resource-metadata
$ ln -s $(pwd)/hub /tmp/hub

# Now modify the trigger file and refresh ui
```
