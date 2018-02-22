apiVersion: v1
kind: Service
metadata:
  annotations:
    git-commit-hash: 9b16688a8cd4304ef3dc84701f7d34e3342c0465
  name: foocorp
  namespace: default
spec:
  selector:
    serviceName: foocorp
