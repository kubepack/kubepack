# UI Editor APIs

- GET "/bundleview"

- POST "/bundleview/orders", v1alpha1.BundleView{}

  - API PATH CHANGED from /deploy/orders -> /bundleview/orders

- GET "/packageview"

`http://localhost:4000/packageview?url=https://bundles.byte.builders/ui/&name=mongodb-editor-options&version=v0.1.0`

- POST "/packageview/orders"

`curl -X POST -H "Content-Type: application/json" -d @./artifacts/mongodb-editor/packageview_chart_order.json http://localhost:4000/packageview/orders?url=https://bundles.byte.builders/ui/&name=mongodb-editor-options&version=v0.1.0`

`$ {"kind":"Order","apiVersion":"kubepack.com/v1alpha1","metadata":{"name":"mymongo","namespace":"demo","uid":"d96b7440-2fdb-4fab-89b0-81d2b72631f2","creationTimestamp":"2021-01-13T05:36:37Z"},"spec":{"items":[{"chart":{"url":"https://bundles.byte.builders/ui/","name":"mongodb-editor-options","version":"v0.1.0","releaseName":"mymongo","namespace":"demo","valuesFile":"values.yaml","valuesPatch":[{"op":"add","path":"/metadata/release/name","value":"mymongo"},{"op":"add","path":"/metadata/release/namespace","value":"demo"},{"op":"replace","path":"/spec/version","value":"4.3.2"}]}}]},"status":{}}`

- GET "/packageview/files"

`http://localhost:4000/packageview/files?url=https://bundles.byte.builders/ui/&name=mongodb-editor-options&version=v0.1.0`

- GET "/packageview/files/\*"

`http://localhost:4000/packageview/files/templates/app.yaml?url=https://bundles.byte.builders/ui/&name=mongodb-editor-options&version=v0.1.0`

`http://localhost:4000/packageview/files/values.yaml?url=https://bundles.byte.builders/ui/&name=mongodb-editor-options&version=v0.1.0&format=json`

- PUT "/editor/model" (Initial Model)

`curl -X PUT -H "Content-Type: application/json" -d @./artifacts/mongodb-editor/mongodb_options_model.json http://localhost:4000/editor/model > ./artifacts/mongodb-editor/mongodb_editor_model.json`

- PUT "/editor/manifest" (Preview API)

`curl -X PUT -H "Content-Type: application/json" -d @./artifacts/mongodb-editor/mongodb_editor_model.json http://localhost:4000/editor/manifest > ./artifacts/mongodb-editor/mongodb_editor_manifest.yaml`

- PUT "/editor/resources" (Preview API)

`curl -X PUT -H "Content-Type: application/json" -d @./artifacts/mongodb-editor/mongodb_editor_model.json http://localhost:4000/editor/resources?skipCRDs=true | jq '.' > ./artifacts/mongodb-editor/mongodb_editor_resources.json`

- POST "/deploy/orders"

- GET "/deploy/orders/:id/render/manifest"

http://localhost:4000/deploy/orders/5902b772-319c-40c1-b260-68d81b7864fd/render/manifest

- GET "/deploy/orders/:id/render/resources"
  - Query parameter: skipCRDs=true

http://localhost:4000/deploy/orders/5902b772-319c-40c1-b260-68d81b7864fd/render/resources?skipCRDs=true

- PUT "/clusters/:cluster/editor" (apply/install/update app API)

`curl -X PUT -H "Content-Type: application/json" -d @./artifacts/mongodb-editor/mongodb_editor_model.json  http://localhost:4000/clusters/my_cluster/editor?installCRDs=true`

- DELETE "/clusters/:cluster/editor/namespaces/:namespace/releases/:releaseName" (Delete app api)

`curl -X DELETE -H "Content-Type: application/json" http://localhost:4000/clusters/my_cluster/editor/namespaces/demo/releases/mymongo`

## UI Edit mode

- PUT "/clusters/my_cluster/editor/model"

`curl -X PUT -H "Content-Type: application/json" -d @./artifacts/mongodb-editor/mongodb_editor_model.json  http://localhost:4000/clusters/my_cluster/editor/model`


- GET "/clusters/:cluster/editor/manifest"
  - redundant apis
  - can be replaced by getting the model, then using the /editor apis

`curl -X PUT -H "Content-Type: application/json" -d @./artifacts/mongodb-editor/mongodb_editor_model.json  http://localhost:4000/clusters/my_cluster/editor/manifest`


- GET "/clusters/:cluster/editor/resources"
  - redundant apis
  - can be replaced by getting the model, then using the /editor apis

`curl -X PUT -H "Content-Type: application/json" -d @./artifacts/mongodb-editor/mongodb_editor_model.json  http://localhost:4000/clusters/my_cluster/editor/resources`
