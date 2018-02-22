apiVersion: v1
kind: Service
metadata:
  annotations:
    git-commit-hash: 14f51ecf7e626588776358eb4bc5add7926642eb
  name: foocorp2
  namespace: default
spec:
  selector:
    serviceName: foocorp2
