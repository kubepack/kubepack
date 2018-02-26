apiVersion: v1
kind: Service
metadata:
  annotations:
    git-commit-hash: d8b4cd1e8e3ef2db940d1daff9695dad42cc005b
  name: foocorp
  namespace: kube-system
spec:
  selector:
    serviceName: foocorp
