# Scenario where KEDA scales to zero after initial activation

```bash
kubectl create ns test-zero
namespace/test-zero created

kubectl apply -f deployment-zero.yaml -n test-zero
deployment.apps/sample-app created
service/sample-app created

kubectl apply -f scale-m.yaml -n test-zero
scaledobject.keda.sh/sample created

kubectl get po -n test-zero
NAME                          READY   STATUS    RESTARTS   AGE
sample-app-78f7b69557-kbs72   1/1     Running   0          52s

kubectl  get hpa -n test-zero
NAME              REFERENCE               TARGETS     MINPODS   MAXPODS   REPLICAS   AGE
keda-hpa-sample   Deployment/sample-app   0/2 (avg)   1         1         1          99s

# initially we are in non active state as there are no metrics
kucebtl  get scaledobjects -n test-zero
NAME     SCALETARGETKIND      SCALETARGETNAME   MIN   MAX   TRIGGERS     AUTHENTICATION   READY   ACTIVE   FALLBACK   PAUSED    AGE
sample   apps/v1.Deployment   sample-app        1     1     prometheus                    True    False    False      Unknown   105s

minikube service list
|-----------------|----------------------------------------------------|--------------|-----------------------------|
|    NAMESPACE    |                        NAME                        | TARGET PORT  |             URL             |
|-----------------|----------------------------------------------------|--------------|-----------------------------|
...
| test-zero       | sample-app                                         | http/8090    | http://192.168.39.215:32641 |
...
|-----------------|----------------------------------------------------|--------------|-----------------------------|

curl http://192.168.39.215:32641
Hello!

curl http://192.168.39.215:32641/metrics
# HELP go_gc_duration_seconds A summary of the pause duration of garbage collection cycles.
# TYPE go_gc_duration_seconds summary
...
my_gauge 0
...
# HELP promhttp_metric_handler_requests_total Total number of scrapes by HTTP status code.
# TYPE promhttp_metric_handler_requests_total counter
promhttp_metric_handler_requests_total{code="200"} 0
promhttp_metric_handler_requests_total{code="500"} 0
promhttp_metric_handler_requests_total{code="503"} 0

# start emitting 5 for the gauge metric
curl http://192.168.39.215:32641/toggle

kubectl get hpa -n test-zero
NAME              REFERENCE               TARGETS     MINPODS   MAXPODS   REPLICAS   AGE
keda-hpa-sample   Deployment/sample-app   5/2 (avg)   1         1         1          3m32s

kubectl get scaledobject -n test-zero
NAME     SCALETARGETKIND      SCALETARGETNAME   MIN   MAX   TRIGGERS     AUTHENTICATION   READY   ACTIVE   FALLBACK   PAUSED    AGE
sample   apps/v1.Deployment   sample-app        1     1     prometheus                    True    True     False      Unknown   3m38s

kubectl get po -n test-zero
NAME                          READY   STATUS    RESTARTS   AGE
sample-app-78f7b69557-kbs72   1/1     Running   0          4m25s

# set minReplicaCount=1
scaledobject.keda.sh/sample configured

kubectl get scaledobject -n test-zero
NAME     SCALETARGETKIND      SCALETARGETNAME   MIN   MAX   TRIGGERS     AUTHENTICATION   READY   ACTIVE   FALLBACK   PAUSED    AGE
sample   apps/v1.Deployment   sample-app        0     1     prometheus                    True    True     False      Unknown   4m30s

kubectl get hpa -n test-zero
NAME              REFERENCE               TARGETS     MINPODS   MAXPODS   REPLICAS   AGE
keda-hpa-sample   Deployment/sample-app   5/2 (avg)   1         1         1          4m35s

# start emitting 0 for the gauge metric
curl http://192.168.39.215:32641/toggle

kubectl get scaledobject -n test-zero
NAME     SCALETARGETKIND      SCALETARGETNAME   MIN   MAX   TRIGGERS     AUTHENTICATION   READY   ACTIVE   FALLBACK   PAUSED    AGE
sample   apps/v1.Deployment   sample-app        0     1     prometheus                    True    True     False      Unknown   4m47s

kubectl get po -n test-zero
NAME                          READY   STATUS    RESTARTS   AGE
sample-app-78f7b69557-kbs72   1/1     Running   0          5m28s

# ScaledObject is set to not active
kubectl get scaledobject -n test-zero
NAME     SCALETARGETKIND      SCALETARGETNAME   MIN   MAX   TRIGGERS     AUTHENTICATION   READY   ACTIVE   FALLBACK   PAUSED    AGE
sample   apps/v1.Deployment   sample-app        0     1     prometheus                    True    False    False      Unknown   5m20s

# after 30secs of cooldown period no instance is running
kubectl get po -n test-zero
No resources found in test-zero namespace.

kubectl get hpa -n test-zero
NAME              REFERENCE               TARGETS     MINPODS   MAXPODS   REPLICAS   AGE
keda-hpa-sample   Deployment/sample-app   0/2 (avg)   1         1         1          5m27s

kubectl get hpa -n test-zero
NAME              REFERENCE               TARGETS             MINPODS   MAXPODS   REPLICAS   AGE
keda-hpa-sample   Deployment/sample-app   <unknown>/2 (avg)   1         1         0          26m
```