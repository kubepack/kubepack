apiVersion: v1
kind: Service
metadata:
  annotations:
    git-commit-hash: e01fb69ae2acf4bc920c2ada8415ff0df0a46ef9
  name: foocorp
  namespace: kube-system
spec:
  selector:
    serviceName: foocorp
