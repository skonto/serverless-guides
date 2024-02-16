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
kubectl apply -f https://github.com/knative/net-kourier/releases/download/knative-v1.13.0/kourier.yaml
kubectl apply -f https://github.com/knative/serving/releases/download/knative-v1.13.1/serving-hpa.yaml

kubectl patch configmap/config-network \
  -n knative-serving \
  --type merge \
  -p '{"data":{"ingress.class":"kourier.ingress.networking.knative.dev"}}'

kubectl patch configmap/config-domain \
      --namespace knative-serving \
      --type merge \
      --patch '{"data":{"example.com":""}}'
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
  
helm  install prometheus prometheus-community/kube-prometheus-stack -n default -f values1.yaml

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

Note here that metrics of the form _total are re-written based on some default rules in the prometheus adapter's cm:

```
- seriesQuery: '{namespace!="",__name__!~"^container_.*"}'
      seriesFilters:
      - isNot: .*_seconds_total
      resources:
        template: <<.Resource>>
      name:
        matches: ^(.*)_total$
        as: ""
      metricsQuery: sum(rate(<<.Series>>{<<.LabelMatchers>>}[5m])) by (<<.GroupBy>>)
```
You can change the rules by modifying the cm: `kubectl edit cm my-release-ad-prometheus-adapter`.
For more check here: https://github.com/kubernetes-sigs/prometheus-adapter/blob/master/docs/config-walkthrough.md.

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

# make sure you see the metric registered
kubectl get --raw "/apis/custom.metrics.k8s.io/v1beta2/namespaces/test/pods/*/http_requests"
{"kind":"MetricValueList","apiVersion":"custom.metrics.k8s.io/v1beta2","metadata":{},"items":[{"describedObject":{"kind":"Pod","namespace":"test","name":"metrics-test-00001-deployment-6d6f98b8bd-fmsrm","apiVersion":"/v1"},"metric":{"name":"http_requests","selector":null},"timestamp":"2024-02-16T13:19:30Z","value":"19m"}]}


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

Note that an important difference of Knative compared to Keda is that initially Knative will scale the deployment from 0 to 1 in order to make sure it works. 
This also makes metrics available to Knative since it scrapes the pods directly. With KEDA since metrics for this scenario here come from the pod itself (not an external service) a pod instance must exist to emmit metrics in the first place.
Then KEDA can decide when to set a scaledObject as active. Thus, if we start with minReplicaCount=0 Keda will not be able to scale the deployment automatically.
In order to set the scaledObject as active we need to initial start with minReplicaCount=1, then scaledObject will become active (assuming certain criteria are met such as metric goes beyond the threshold) and then  if the custom metric gets to zero it will scale down to zero the deployment and the scaledObject will become inactive.
For more check the example [here](./keda/README.md).

# Knative HPA on OCP

## Install Serverless

```
oc apply -f serverless-kourier.yaml
oc wait --for=condition=ready pod -l name=knative-openshift -n openshift-serverless --timeout=300s
oc wait --for=condition=ready pod -l name=knative-openshift-ingress -n openshift-serverless --timeout=300s
oc wait --for=condition=ready pod -l name=knative-operator -n openshift-serverless --timeout=300s
oc apply -f knativeserving-kourier.yaml
```

## Install Prometheus and Prometheus Adapter

a) Install Prometheus Operator eg. from operatohub 

b) Install Prometheus Instance using for example:

```
apiVersion: monitoring.coreos.com/v1
kind: Prometheus
metadata:
  name: app-monitor
  labels:
    prometheus: k8s
spec:
  replicas: 1
  serviceAccountName: prometheus-k8s
  securityContext: {}
  ruleSelector: {}
  serviceMonitorNamespaceSelector: {}
  serviceMonitorSelector: {}
  alerting:
    alertmanagers:
      - namespace: openshift-monitoring
        name: alertmanager-main
        port: web
---
apiVersion: v1
kind: Service
metadata:
  labels:
    operated-prometheus: "true"
  name: prometheus-operated
  namespace: default
spec:
  clusterIP: None
  clusterIPs:
  - None
  internalTrafficPolicy: Cluster
  ipFamilies:
  - IPv4
  ipFamilyPolicy: SingleStack
  ports:
  - name: web
    port: 9090
    protocol: TCP
    targetPort: web
  selector:
    app.kubernetes.io/name: prometheus
  sessionAffinity: None
  type: ClusterIP
```

Then install the Prometheus adapter using as values file:
```
prometheus:
  url: http://prometheus-operated.default.svc
 
```

Follow the examples above.

# Knative KEDA Integration

Please take a look at https://github.com/skonto/serving-keda.