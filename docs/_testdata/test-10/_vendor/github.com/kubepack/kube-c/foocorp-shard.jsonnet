apiVersion: v1
kind: Service
metadata:
  name: foocorp
  namespace: kube-system
spec:
  selector:
    serviceName: foocorp
