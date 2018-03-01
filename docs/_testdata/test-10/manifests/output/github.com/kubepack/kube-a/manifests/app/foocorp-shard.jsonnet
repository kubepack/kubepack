apiVersion: v1
kind: Service
metadata:
  annotations:
    git-commit-hash: 0c4294b79d7b8ba8851526399007238c459079e6
  name: foocorp2
  namespace: default
spec:
  selector:
    serviceName: foocorp2
