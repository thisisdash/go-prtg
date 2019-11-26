# Sync PRTG devices based on k8s ingress objects

This example shows how you can create/update devices in
PRTG based on kubernetes ingress objects.

Configuration is done on the `Syncer` object, this example expects an ingress of the format:

```
apiVersion: networking.k8s.io/v1beta1
kind: Ingress
metadata:
  name: test
  annotations:
    "prometheus.io/path": "/healthz"
spec:
  rules:
  - host: myapplication.example.com
    http:
      paths:
      - backend:
          serviceName: nginx
          servicePort: 80
```