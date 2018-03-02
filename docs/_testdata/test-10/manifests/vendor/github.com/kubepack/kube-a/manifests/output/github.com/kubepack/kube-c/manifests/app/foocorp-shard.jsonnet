apiVersion: v1
kind: Service
metadata:
  annotations:
    git-commit-hash: 9476d358a28fc5b666b21c89d710bd571bf1c1ca
  name: foocorp1
  namespace: default
spec:
  selector:
    serviceName: foocorp
