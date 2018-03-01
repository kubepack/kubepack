apiVersion: v1
kind: Service
metadata:
  annotations:
    git-commit-hash: 21219163dbc1707e9e670ddc62cbe108b4f672ff
  name: foocorp1
  namespace: default
spec:
  selector:
    serviceName: foocorp
