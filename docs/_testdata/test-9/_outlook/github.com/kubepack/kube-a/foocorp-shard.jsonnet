apiVersion: v1
kind: Service
metadata:
  annotations:
    git-commit-hash: e69be3069044bf99611ca7fe07d26515b3a1ead1
  name: foocorp
  namespace: default
spec:
  selector:
    serviceName: foocorp
