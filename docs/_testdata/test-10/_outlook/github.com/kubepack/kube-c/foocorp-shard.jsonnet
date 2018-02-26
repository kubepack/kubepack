apiVersion: v1
kind: Service
metadata:
  annotations:
    git-commit-hash: f9f5aa2e1e4d8c8868b3d9a69b484b9c75946912
  name: foocorp
  namespace: kube-system
spec:
  selector:
    serviceName: foocorp
