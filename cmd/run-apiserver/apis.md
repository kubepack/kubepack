# UI Editor APIs

- GET "/bundleview"

- POST "/bundleview/orders", v1alpha1.BundleView{}

  - API PATH CHANGED from /deploy/orders -> /bundleview/orders

- GET "/packageview"

`http://localhost:4000/packageview?url=https://bundles.byte.builders/ui/&name=mongodb-editor-options&version=v0.1.0`

- POST "/packageview/orders"

`curl -X POST -H "Content-Type: application/json" \ -d @./artifacts/mongodb-editor/packageview_chart_order.json \ http://localhost:4000/packageview/orders?url=https://bundles.byte.builders/ui/&name=mongodb-editor-options&version=v0.1.0`

`$ {"kind":"Order","apiVersion":"kubepack.com/v1alpha1","metadata":{"name":"mymongo","namespace":"demo","uid":"d96b7440-2fdb-4fab-89b0-81d2b72631f2","creationTimestamp":"2021-01-13T05:36:37Z"},"spec":{"items":[{"chart":{"url":"https://bundles.byte.builders/ui/","name":"mongodb-editor-options","version":"v0.1.0","releaseName":"mymongo","namespace":"demo","valuesFile":"values.yaml","valuesPatch":[{"op":"add","path":"/metadata/release/name","value":"mymongo"},{"op":"add","path":"/metadata/release/namespace","value":"demo"},{"op":"replace","path":"/spec/version","value":"4.3.2"}]}}]},"status":{}}`

- GET "/packageview/files"

`http://localhost:4000/packageview/files?url=https://bundles.byte.builders/ui/&name=mongodb-editor-options&version=v0.1.0`

- GET "/packageview/files/\*"

`http://localhost:4000/packageview/files/templates/app.yaml?url=https://bundles.byte.builders/ui/&name=mongodb-editor-options&version=v0.1.0`

- POST "/editor/:group/:version/:resource/model" (Initial Model)

`curl -X POST -H "Content-Type: application/json" \ -d @./artifacts/mongodb-editor/mongodb_options_model.json \ http://localhost:4000/editor/kubedb.com/v1alpha2/mongodbs/model > ./artifacts/mongodb-editor/mongodb_editor_model.json`

- GET "/editor/:group/:version/:resource/manifest"

`curl -X POST -H "Content-Type: application/json" \ -d @./artifacts/mongodb-editor/mongodb_editor_model.json \ http://localhost:4000/editor/kubedb.com/v1alpha2/mongodbs/mymongo/namespaces/demo/manifest > ./artifacts/mongodb-editor/mongodb_editor_manifest.yaml`

- GET "/editor/:group/:version/:resource/resources"

`curl -X POST -H "Content-Type: application/json" \ -d @./artifacts/mongodb-editor/mongodb_editor_model.json \ http://localhost:4000/editor/kubedb.com/v1alpha2/mongodbs/mymongo/namespaces/demo/resources?skipCRDs=true | jq '.' > ./artifacts/mongodb-editor/mongodb_editor_resources.json`

- POST "/deploy/orders", v1alpha1.Order{}

- GET "/deploy/orders/:id/render/manifest"

- GET "/deploy/orders/:id/render/resources"

- PUT "/clusters/:cluster/editor/:group/:version/namespaces/:namespace/:resource/:releaseName"

`curl -X PUT -H "Content-Type: application/json" \ -d @./artifacts/mongodb-editor/mongodb_editor_model.json \ http://localhost:4000/clusters/my_cluster/editor/kubedb.com/v1alpha2/namespaces/demo/mongodbs/mymongo`

- DELETE "/clusters/:cluster/editor/namespaces/:namespace/releases/:releaseName"

`curl -X DELETE -H "Content-Type: application/json" \ http://localhost:4000/clusters/my_cluster/editor/namespaces/demo/releases/mymongo`

- GET "/clusters/:cluster/editor/:group/:version/namespaces/:namespace/:resource/:releaseName/model"

`http://localhost:4000/clusters/my_cluster/editor/kubedb.com/v1alpha2/namespaces/demo/mongodbs/mymongo/model`

- GET "/clusters/:cluster/editor/:group/:version/namespaces/:namespace/:resource/:releaseName/manifest"
  - redundant apis
  - can be replaced by getting the model, then using the /editor apis

`http://localhost:4000/clusters/my_cluster/editor/kubedb.com/v1alpha2/namespaces/demo/mongodbs/mymongo/manifest`

- GET "/clusters/:cluster/editor/:group/:version/namespaces/:namespace/:resource/:releaseName/resources"
  - redundant apis
  - can be replaced by getting the model, then using the /editor apis

`http://localhost:4000/clusters/my_cluster/editor/kubedb.com/v1alpha2/namespaces/demo/mongodbs/mymongo/resources`
