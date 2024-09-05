# Instructions


```bash
# run locally on Minikube
kubectl apply -f ksvc.yaml

#grab the Kourier gateway ip with `minikube service list` 
go run cmd/ws/ws_client.go websocket-server.default.example.com 192.168.39.83:30950 0
```
