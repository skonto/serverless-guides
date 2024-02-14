# Knative HPA on Minikube

## Install Minikube and Knative Serving

```
minikube delete

MEMORY=${MEMORY:-40000}
CPUS=${CPUS:-6}

EXTRA_CONFIG="apiserver.enable-admission-plugins=\
LimitRanger,\
NamespaceExists,\
NamespaceLifecycle,\
ResourceQuota,\
ServiceAccount,\
DefaultStorageClass,\
MutatingAdmissionWebhook"

minikube start --driver=kvm2 --memory=$MEMORY --cpus=$CPUS \
  --kubernetes-version=v1.27.6 \
  --disk-size=30g \
  --extra-config="$EXTRA_CONFIG" \
  --extra-config=kubelet.authentication-token-webhook=true

kubectl apply -f https://github.com/knative/serving/releases/download/knative-v1.13.1/serving-crds.yaml
kubectl apply -f https://github.com/knative/serving/releases/download/knative-v1.13.1/serving-core.yaml
kubectl apply -f https://github.com/knative/net-kourier/releases/download/knative-v1.13.1/kourier.yaml
kubectl apply -f https://github.com/knative/serving/releases/download/knative-v1.13.1/serving-hpa.yaml
```

## Install Prometheus and Prometheus Adapter

```bash
cat values1.yaml:
kube-state-metrics:
metricLabelsAllowlist:
- pods=[*]
- deployments=[app.kubernetes.io/name,app.kubernetes.io/component,app.kubernetes.io/instance]
  prometheus:
  prometheusSpec:
  serviceMonitorSelectorNilUsesHelmValues: false
  podMonitorSelectorNilUsesHelmValues: false
  grafana:
  sidecar:
  dashboards:
  enabled: true
  searchNamespace: ALL
  
helm  install prometheus prometheus-community/kube-prometheus-stack -n default -f values.yaml

cat values2.yaml:
prometheus:
url: http://prometheus-kube-prometheus-prometheus.default.svc

helm install my-release-ad prometheus-community/prometheus-adapter -f values2.yaml
```

## Configure the metrics aggregation

Apply
```bash
kubectl apply -f - <<EOF
apiVersion: apiregistration.k8s.io/v1
kind: APIService
metadata:
  name: v1beta2.custom.metrics.k8s.io
spec:
  group: custom.metrics.k8s.io
  groupPriorityMinimum: 100
  insecureSkipTLSVerify: true
  service:
    name: my-release-ad-prometheus-adapter
    namespace: default
  version: v1beta2
  versionPriority: 100
EOF
```

Based on walkthrough here: https://github.com/kubernetes-sigs/prometheus-adapter/blob/master/docs/walkthrough.md.

For a K8s dep:

```bash
kubectl apply -f deployment.yaml -n test
```

Applying this on minikube in test ns:

```bash
minikube service list | grep kourier

curl http://192.168.39.169:30550
Hello! My name is sample-app-5499d69c59-vqst4. I have served 1267 requests so far.

kubectl get hpa -n test
NAME         REFERENCE               TARGETS    MINPODS   MAXPODS   REPLICAS   AGE
sample-app   Deployment/sample-app   37m/500m   1         10        1          126m
```

Doing the same with a Knative ksvc, apply the following first in serving-test ns:

```bash
kubectl apply -f ksvc-hpa.yaml -n test

curl  -H "Host: metrics-test.serving-test.example.com" http://192.168.39.169:31534/metrics
# HELP http_requests_total The amount of requests served by the server in total
# TYPE http_requests_total counter
http_requests_total 4157

kubectl get hpa -n serving-test
NAME                 REFERENCE                                  TARGETS   MINPODS   MAXPODS   REPLICAS   AGE
metrics-test-00001   Deployment/metrics-test-00001-deployment   33m/10    1         10        1          52m

# Create some load (repeat a few times): 
for i in {1..1000}; do curl  -H "Host: metrics-test.serving-test.example.com" http://192.168.39.169:31534/metrics; done

kubectl get hpa -n serving-test
NAME                 REFERENCE                                  TARGETS     MINPODS   MAXPODS   REPLICAS   AGE
metrics-test-00001   Deployment/metrics-test-00001-deployment   11465m/10   1         10        3          58m

kubectl get po -n serving-test --watch
NAME                                             READY   STATUS    RESTARTS   AGE
metrics-test-00001-deployment-59579b768d-665wx   2/2     Running   0          88m
metrics-test-00001-deployment-59579b768d-d7kqj   2/2     Running   0          47s
metrics-test-00001-deployment-59579b768d-r8dfm   0/2     Pending   0          0s
metrics-test-00001-deployment-59579b768d-r8dfm   0/2     Pending   0          0s
metrics-test-00001-deployment-59579b768d-r8dfm   0/2     ContainerCreating   0          0s
metrics-test-00001-deployment-59579b768d-r8dfm   1/2     Running             0          3s
metrics-test-00001-deployment-59579b768d-r8dfm   2/2     Running             0          4s
```

# Knative HPA on OCP

## Install Serverless

oc apply -f serverless.yaml

oc wait --for=condition=ready pod -l name=knative-openshift -n openshift-serverless --timeout=300s
oc wait --for=condition=ready pod -l name=knative-openshift-ingress -n openshift-serverless --timeout=300s
oc wait --for=condition=ready pod -l name=knative-operator -n openshift-serverless --timeout=300s

oc apply -f knativeserving-kourier.yaml


## Install Prometheus and Prometheus Adapter

TBD


# Knative KEDA Integration

Please take a look at https://github.com/skonto/serving-keda.