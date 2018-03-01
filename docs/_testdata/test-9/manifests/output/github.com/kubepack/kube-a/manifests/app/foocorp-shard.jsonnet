apiVersion: v1
kind: Service
metadata:
  annotations:
    git-commit-hash: 8bc0f6d2243b194f970c92078000ee4a4d00325d
  name: foocorp
  namespace: default
spec:
  selector:
    serviceName: foocorp
