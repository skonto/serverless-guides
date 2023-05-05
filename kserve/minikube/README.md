# Minikube Setup

```
# This depends on the host machine, here being probably more generous than we should
MEMORY=${MEMORY:-16000}
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
  --kubernetes-version=v1.26.1 \
  --disk-size=30g \
  --extra-config="$EXTRA_CONFIG" \
  --extra-config=kubelet.authentication-token-webhook=true
```

```
# In a seperate terminal run

minikube tunnel
```

# Istio Setup

```
$ curl -L https://istio.io/downloadIstio | ISTIO_VERSION=1.17.1 TARGET_ARCH=x86_64 sh -

$ export PATH="$PATH:.../istio-1.17.1/bin"
```

```
$ istioctl manifest generate > ./generated-manifest.yaml

$ istioctl profile dump default --config-path values.global.proxy.resources
limits:
  cpu: 2000m
  memory: 1024Mi
requests:
  cpu: 100m
  memory: 128Mi
```
```
$ istioctl install -y
```
```
$ istioctl verify-install -f ./generated-manifest.yaml
✔ CustomResourceDefinition: authorizationpolicies.security.istio.io.istio-system checked successfully
..
Checked 15 custom resource definitions
Checked 2 Istio Deployments
✔ Istio is installed and verified successfully


```

# Knative Setup

```
$ kubectl apply -f https://github.com/knative/serving/releases/download/knative-v1.9.2/serving-crds.yaml
$ kubectl apply -f https://github.com/knative/serving/releases/download/knative-v1.9.2/serving-core.yaml
$ kubectl apply -f https://github.com/knative/net-istio/releases/download/knative-v1.9.2/net-istio.yaml
```

```
$ kubectl patch configmap/config-domain \
  --namespace knative-serving \
  --type merge \
  --patch '{"data":{"example.com":""}}'

# This restricts (among others) app monitoring, check istio docs for more on how to setup Prometheus.

$ kubectl apply -f - <<EOF
apiVersion: security.istio.io/v1beta1
kind: PeerAuthentication
metadata:
  name: default
  namespace: istio-system
spec:
  mtls:
    mode: STRICT
EOF
```

```
$ kubectl label namespace knative-serving istio-injection=enabled
```
```
# Alternatively to avoid adding all pods in the mesh injection can be disabled
# at the pod level by editing only the activator, autoscaler deployments and by adding the annotation:
spec:
  template:
    metadata:
      annotations:
        sidecar.istio.io/inject: "false"
For more check: https://stackoverflow.com/questions/65807748/disable-istio-sidecar-injection-to-the-job-pod

Also it might be useful to add (depending on your setup eg disabled globally):
      "sidecar.istio.io/rewriteAppHTTPProbers": "true"

For more check here: https://istio.io/latest/docs/ops/configuration/mesh/app-health-check/

```

```
# restart pods to get side-car injected

kubectl delete pods --all -n knative-serving

kubectl get pods -n knative-serving
NAME                                     READY   STATUS    RESTARTS   AGE
activator-556dc846dd-fnl48               2/2     Running   0          57m
autoscaler-57c75556d6-bpcf4              2/2     Running   0          57m
controller-67d4f547cb-tb62c              2/2     Running   0          57m
domain-mapping-6646494575-dhtdx          2/2     Running   0          57m
domainmapping-webhook-5c89cfc845-ms8cg   2/2     Running   0          57m
net-istio-controller-8659b4dc74-k6rrv    1/1     Running   0          57m
net-istio-webhook-55b895d9b8-4rg9n       2/2     Running   0          57m
webhook-6f67fd848c-r7fn8                 2/2     Running   0          57m
```

## Verify Knative with Istio setup

```
kubectl create ns test
kubectl label namespace test istio-injection=enabled

kubectl apply -n test -f - <<EOF
apiVersion: serving.knative.dev/v1
kind: Service
metadata:
  name: helloworld-go
spec:
  template:
    spec:
      containers:
        - image: gcr.io/knative-samples/helloworld-go
          env:
            - name: TARGET
              value: "Go Sample v1"
EOF

curl -H "Host: helloworld-go.test.example.com" http://192.168.39.224:31940
Hello Go Sample v1!
```

# KServe

```
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.10.1/cert-manager.yaml

kubectl get pods -n cert-manager
NAME                                       READY   STATUS    RESTARTS   AGE
cert-manager-7fb78674d7-vlfkx              1/1     Running   0          36m
cert-manager-cainjector-5dfc946d84-j8pkt   1/1     Running   0          36m
cert-manager-webhook-8744b7588-82fln       1/1     Running   0          36m


kubectl apply -f  https://github.com/kserve/kserve/releases/download/v0.10.1/kserve.yaml
kubectl apply -f  https://github.com/kserve/kserve/releases/download/v0.10.1/kserve-runtimes.yaml
```

```
kubectl get pods -n kserve
NAME                                         READY   STATUS    RESTARTS   AGE
kserve-controller-manager-59867d7d9f-7fnml   2/2     Running   0          34m
```

## Verify Kserve

```
kubectl apply -n test -f - <<EOF
apiVersion: "serving.kserve.io/v1beta1"
kind: "InferenceService"
metadata:
  name: "pmml-demo"
spec:
  predictor:
    pmml:
      storageUri: gs://kfserving-examples/models/pmml
EOF
```
```
cat <<EOF > "./pmml.json"
{
  "instances": [
    [5.1, 3.5, 1.4, 0.2]
  ]
}
EOF

```

```
kubectl get pods -n test
NAME                                                            READY   STATUS    RESTARTS   AGE
pmml-demo-predictor-default-00001-deployment-5b57887d98-vlpfl   3/3     Running   0          37m
```

```
curl -H "Host: pmml-demo-predictor-default.test.example.com" http://192.168.39.224:31940/v1/models/pmml-demo:predict -d @./pmml.json
{"predictions": [{"Species": "setosa", "Probability_setosa": 1.0, "Probability_versicolor": 0.0, "Probability_virginica": 0.0, "Node_Id": "2"}]}[

```

```
kubectl get inferenceservices.serving.kserve.io
NAME        URL                                 READY   PREV   LATEST   PREVROLLEDOUTREVISION   LATESTREADYREVISION                 AGE
pmml-demo   http://pmml-demo.test.example.com   True           100                              pmml-demo-predictor-default-00001   124m

kubectl get inferenceservices.serving.kserve.io -ojson
{
    "apiVersion": "v1",
    "items": [
        {
            "apiVersion": "serving.kserve.io/v1beta1",
            "kind": "InferenceService",
            "metadata": {
                "annotations": {
                    "kubectl.kubernetes.io/last-applied-configuration": "{\"apiVersion\":\"serving.kserve.io/v1beta1\",\"kind\":\"InferenceService\",\"metadata\":{\"annotations\":{},\"name\":\"pmml-demo\",\"namespace\":\"test\"},\"spec\":{\"predictor\":{\"pmml\":{\"storageUri\":\"gs://kfserving-examples/models/pmml\"}}}}\n"
                },
                "creationTimestamp": "2023-03-21T18:54:11Z",
                "finalizers": [
                    "inferenceservice.finalizers"
                ],
                "generation": 1,
                "name": "pmml-demo",
                "namespace": "test",
                "resourceVersion": "12522",
                "uid": "f99324ed-9385-432e-a863-8af5ea98b820"
            },
            "spec": {
                "predictor": {
                    "model": {
                        "modelFormat": {
                            "name": "pmml"
                        },
                        "name": "",
                        "resources": {},
                        "storageUri": "gs://kfserving-examples/models/pmml"
                    }
                }
            },
            "status": {
                "address": {
                    "url": "http://pmml-demo.test.svc.cluster.local/v1/models/pmml-demo:predict"
                },
                "components": {
                    "predictor": {
                        "address": {
                            "url": "http://pmml-demo-predictor-default.test.svc.cluster.local"
                        },
                        "latestCreatedRevision": "pmml-demo-predictor-default-00001",
                        "latestReadyRevision": "pmml-demo-predictor-default-00001",
                        "latestRolledoutRevision": "pmml-demo-predictor-default-00001",
                        "traffic": [
                            {
                                "latestRevision": true,
                                "percent": 100,
                                "revisionName": "pmml-demo-predictor-default-00001"
                            }
                        ],
                        "url": "http://pmml-demo-predictor-default.test.example.com"
                    }
                },
                "conditions": [
                    {
                        "lastTransitionTime": "2023-03-21T19:06:09Z",
                        "status": "True",
                        "type": "IngressReady"
                    },
                    {
                        "lastTransitionTime": "2023-03-21T19:06:09Z",
                        "severity": "Info",
                        "status": "True",
                        "type": "PredictorConfigurationReady"
                    },
                    {
                        "lastTransitionTime": "2023-03-21T19:06:09Z",
                        "status": "True",
                        "type": "PredictorReady"
                    },
                    {
                        "lastTransitionTime": "2023-03-21T19:06:09Z",
                        "severity": "Info",
                        "status": "True",
                        "type": "PredictorRouteReady"
                    },
                    {
                        "lastTransitionTime": "2023-03-21T19:06:09Z",
                        "status": "True",
                        "type": "Ready"
                    }
                ],
                "modelStatus": {
                    "copies": {
                        "failedCopies": 0,
                        "totalCopies": 1
                    },
                    "states": {
                        "activeModelState": "Loaded",
                        "targetModelState": "Loaded"
                    },
                    "transitionStatus": "UpToDate"
                },
                "observedGeneration": 1,
                "url": "http://pmml-demo.test.example.com"
            }
        }
    ],
    "kind": "List",
    "metadata": {
        "resourceVersion": ""
    }
}

kubectl get sks -n test
NAME                                MODE    ACTIVATORS   SERVICENAME                         PRIVATESERVICENAME                          READY     REASON
helloworld-go-00001                 Proxy   3            helloworld-go-00001                 helloworld-go-00001-private                 Unknown   NoHealthyBackends
pmml-demo-predictor-default-00001   Proxy   4            pmml-demo-predictor-default-00001   pmml-demo-predictor-default-00001-private   True      

kubectl get services -n test
NAME                                        TYPE           CLUSTER-IP       EXTERNAL-IP                                            PORT(S)                                              AGE
helloworld-go                               ExternalName   <none>           knative-local-gateway.istio-system.svc.cluster.local   80/TCP                                               77m
helloworld-go-00001                         ClusterIP      10.109.202.173   <none>                                                 80/TCP,443/TCP                                       78m
helloworld-go-00001-private                 ClusterIP      10.104.197.190   <none>                                                 80/TCP,443/TCP,9090/TCP,9091/TCP,8022/TCP,8012/TCP   78m
pmml-demo                                   ExternalName   <none>           knative-local-gateway.istio-system.svc.cluster.local   <none>                                               46m
pmml-demo-predictor-default                 ExternalName   <none>           knative-local-gateway.istio-system.svc.cluster.local   80/TCP                                               46m
pmml-demo-predictor-default-00001           ClusterIP      10.97.2.235      <none>                                                 80/TCP,443/TCP                                       57m
pmml-demo-predictor-default-00001-private   ClusterIP      10.97.12.86      <none>                                                 80/TCP,443/TCP,9090/TCP,9091/TCP,8022/TCP,8012/TCP   57m

kubectl get ksvc -n test
NAME                          URL                                                   LATESTCREATED                       LATESTREADY                         READY   REASON
helloworld-go                 http://helloworld-go.test.example.com                 helloworld-go-00001                 helloworld-go-00001                 True    
pmml-demo-predictor-default   http://pmml-demo-predictor-default.test.example.com   pmml-demo-predictor-default-00001   pmml-demo-predictor-default-00001   True    


kubectl get ksvc pmml-demo-predictor-default -n test -ojson
{
    "apiVersion": "serving.knative.dev/v1",
    "kind": "Service",
    "metadata": {
        "annotations": {
            "serving.knative.dev/creator": "system:serviceaccount:kserve:kserve-controller-manager",
            "serving.knative.dev/lastModifier": "system:serviceaccount:kserve:kserve-controller-manager"
        },
        "creationTimestamp": "2023-03-21T18:54:12Z",
        "generation": 1,
        "labels": {
            "component": "predictor",
            "serving.kserve.io/inferenceservice": "pmml-demo"
        },
        "name": "pmml-demo-predictor-default",
        "namespace": "test",
        "ownerReferences": [
            {
                "apiVersion": "serving.kserve.io/v1beta1",
                "blockOwnerDeletion": true,
                "controller": true,
                "kind": "InferenceService",
                "name": "pmml-demo",
                "uid": "f99324ed-9385-432e-a863-8af5ea98b820"
            }
        ],
        "resourceVersion": "12517",
        "uid": "ecdd7dcb-c484-4d86-8a47-56950df8ae5b"
    },
    "spec": {
        "template": {
            "metadata": {
                "annotations": {
                    "autoscaling.knative.dev/class": "kpa.autoscaling.knative.dev",
                    "autoscaling.knative.dev/min-scale": "1",
                    "internal.serving.kserve.io/storage-initializer-sourceuri": "gs://kfserving-examples/models/pmml",
                    "prometheus.kserve.io/path": "/metrics",
                    "prometheus.kserve.io/port": "8080"
                },
                "creationTimestamp": null,
                "labels": {
                    "component": "predictor",
                    "serving.kserve.io/inferenceservice": "pmml-demo"
                }
            },
            "spec": {
                "containerConcurrency": 0,
                "containers": [
                    {
                        "args": [
                            "--model_name=pmml-demo",
                            "--model_dir=/mnt/models",
                            "--http_port=8080"
                        ],
                        "image": "kserve/pmmlserver:v0.10.1",
                        "name": "kserve-container",
                        "readinessProbe": {
                            "successThreshold": 1,
                            "tcpSocket": {
                                "port": 0
                            }
                        },
                        "resources": {
                            "limits": {
                                "cpu": "1",
                                "memory": "2Gi"
                            },
                            "requests": {
                                "cpu": "1",
                                "memory": "2Gi"
                            }
                        }
                    }
                ],
                "enableServiceLinks": false,
                "timeoutSeconds": 300
            }
        },
        "traffic": [
            {
                "latestRevision": true,
                "percent": 100
            }
        ]
    },
    "status": {
        "address": {
            "url": "http://pmml-demo-predictor-default.test.svc.cluster.local"
        },
        "conditions": [
            {
                "lastTransitionTime": "2023-03-21T19:06:08Z",
                "status": "True",
                "type": "ConfigurationsReady"
            },
            {
                "lastTransitionTime": "2023-03-21T19:06:09Z",
                "status": "True",
                "type": "Ready"
            },
            {
                "lastTransitionTime": "2023-03-21T19:06:09Z",
                "status": "True",
                "type": "RoutesReady"
            }
        ],
        "latestCreatedRevisionName": "pmml-demo-predictor-default-00001",
        "latestReadyRevisionName": "pmml-demo-predictor-default-00001",
        "observedGeneration": 1,
        "traffic": [
            {
                "latestRevision": true,
                "percent": 100,
                "revisionName": "pmml-demo-predictor-default-00001"
            }
        ],
        "url": "http://pmml-demo-predictor-default.test.example.com"
    }
}
```

# Observability

## Prometheus, Grafana Setup

A quick way to setup Prometheus is to run:

```
kubectl apply -f https://raw.githubusercontent.com/istio/istio/release-1.17/samples/addons/prometheus.yaml
```

Alternatively you can setup Prometheus Operator that allows more advanced features such as servicemonitor support:

```
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm repo update
helm install prometheus prometheus-community/kube-prometheus-stack -n istio-system -f values.yaml
# values.yaml contains at minimum the configuration below

values.yaml
kube-state-metrics:
  metricLabelsAllowlist:
    - pods=[*]
    - deployments=[app.kubernetes.io/name,app.kubernetes.io/component,app.kubernetes.io/instance]
prometheus:
  prometheusSpec:
    serviceMonitorSelectorNilUsesHelmValues: false
    podMonitorSelectorNilUsesHelmValues: false
```

Edit PeerAuthentication default in istio-system namespace and set it to
PERMISSIVE. Otherwise, you will need to [set certificates for Prometheus](https://istio.io/latest/docs/ops/integrations/prometheus/#tls-settings) to be able to scrape metrics for pods in the service mesh.

Then apply the following podmonitor from the istio samples:

```
kubectl apply -f https://raw.githubusercontent.com/istio/istio/master/samples/addons/extras/prometheus-operator.yaml

```

After setting up Kiali in the next step edit its config in the istio-namespace and add:

```
external-services:
  prometheus:
    url: http://prometheus-operated.istio-system:9090
```
Restart the kiali pod.


```
$ kubectl port-forward svc/prometheus-operated -n istio-system 9090:9090

# Access Prometheus at http://localhost:9090, user:admin, password: prom-operator

$ kubectl port-forward svc/prometheus-grafana  -n istio-system 8080:80

# Access Grafana at http://localhost:8080, user:admin, password: prom-operator

$ kubectl port-forward svc/kiali 20001:20001 -n istio-system

Access Kiali at http://localhost:20001
```

## Kiali Setup


```
$ kubectl apply -f https://raw.githubusercontent.com/istio/istio/release-1.17/samples/addons/kiali.yaml
```

To get the metrics from KServe 8080 port into PRometheus if you have installed PRometheus operator you can use:

```
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  labels:
  name: pmml-demo-predictor-default-sm
spec:
  endpoints:
  - port: model-metrics
    scheme: http
  namespaceSelector: {}
  selector:
    matchLabels:
       name:  pmml-demo-predictor-default-sm
---
apiVersion: v1
kind: Service
metadata:
  labels:
    name:  pmml-demo-predictor-default-sm
  name:  pmml-demo-predictor-default-sm
spec:
  ports:
  - name: model-metrics
    port: 8080
    protocol: TCP
    targetPort: 8080
  selector:
    serving.knative.dev/service: pmml-demo-predictor-default
  type: ClusterIP

```

# GRPC support

Deploy the inference service:

```
kubectl apply -n test -f - <<EOF
apiVersion: serving.kserve.io/v1beta1
kind: InferenceService
metadata:
  name: torchscript-cifar10
spec:
  predictor:
    triton:
      storageUri: gs://kfserving-examples/models/torchscript
      runtimeVersion: 20.10-py3
      ports:
      - containerPort: 9000
        name: h2c
        protocol: TCP
      env:
      - name: OMP_NUM_THREADS
        value: "1"
EOF
```

Run:

```
out_dir="$(mktemp -d /tmp/certs-XXX)"

openssl req -x509 -sha256 -nodes -days 365 -newkey rsa:2048 \
  -subj "/O=Example Inc./CN=Example" \
  -keyout "${out_dir}"/root.key \
  -out "${out_dir}"/root.crt

subdomain="*.test.example.com"
openssl req -nodes -newkey rsa:2048 \
    -subj "/O=Example Inc./CN=Example" \
    -reqexts san \
    -config <(printf "[req]\ndistinguished_name=req\n[san]\nsubjectAltName=DNS:*.%s" "$subdomain") \
    -keyout "${out_dir}"/wildcard.key \
    -out "${out_dir}"/wildcard.csr

openssl x509 -req -days 365 -set_serial 0 \
    -extfile <(printf "subjectAltName=DNS:*.%s" "$subdomain") \
    -CA "${out_dir}"/root.crt \
    -CAkey "${out_dir}"/root.key \
    -in "${out_dir}"/wildcard.csr \
    -out "${out_dir}"/wildcard.crt

kubectl create -n istio-system secret tls wildcard-certs \
    --key="${out_dir}"/wildcard.key \
    --cert="${out_dir}"/wildcard.crt --dry-run=client -o yaml | kubectl apply -f -
```

Update the knative-ingress-gateway:



```
$ kubectl edit  gateways.networking.istio.io  knative-ingress-gateway -n knative-serving

You should have:

apiVersion: networking.istio.io/v1alpha3
kind: Gateway
metadata:
  name: knative-ingress-gateway
  namespace: knative-serving
spec:
  selector:
    istio: ingressgateway
  servers:
    - port:
        number: 443
        name: https
        protocol: HTTPS
      hosts:
        - "*"
      tls:
        mode: SIMPLE
        credentialName: wildcard-certs
```    


```
$  ~/.local/bin/grpcurl  -insecure  -proto ${PROTO_FILE} -authority "torchscript-cifar10.test.example.com" 192.168.39.222:31260   inference.GRPCInferenceService.ServerReady
{
  "ready": true
}

$ ~/.local/bin/grpcurl -insecure -d @ -proto ${PROTO_FILE} -authority "torchscript-cifar10.test.example.com" 192.168.39.222:31260  inference.GRPCInferenceService.ModelInfer <<< $(cat input-grpc.json)
{
  "modelName": "cifar10",
  "modelVersion": "1",
  "outputs": [
    {
      "name": "OUTPUT__0",
      "datatype": "FP32",
      "shape": [
        "1",
        "10"
      ]
    }
  ],
  "rawOutputContents": [
    "vywGwLZLDL7ncgK/dusyQBaAD799KP8/In2QP43As7+UuRk/1uoHwA=="
  ]
}
```


# Canary deployment


```
kubectl apply -n test -f - <<EOF
apiVersion: "serving.kserve.io/v1beta1"
kind: "InferenceService"
metadata:
  name: "sklearn-iris"
spec:
  predictor:
    model:
      modelFormat:
        name: sklearn
      storageUri: "gs://kfserving-examples/models/sklearn/1.0/model"
EOF

cat <<EOF > "./iris-input.json"
{
  "instances": [
    [6.8,  2.8,  4.8,  1.4],
    [6.0,  3.4,  4.5,  1.6]
  ]
}
EOF


kubectl apply -n test -f - <<EOF
apiVersion: "serving.kserve.io/v1beta1"
kind: "InferenceService"
metadata:
  name: "sklearn-iris"
spec:
  predictor:
    canaryTrafficPercent: 10
    model:
      modelFormat:
        name: sklearn
      storageUri: "gs://kfserving-examples/models/sklearn/1.0/model-2"
EOF

```

# Logger dumper


```
kubectl apply -n test -f - <<EOF
apiVersion: serving.knative.dev/v1
kind: Service
metadata:
  name: message-dumper
spec:
  template:
    spec:
      containers:
      - image: gcr.io/knative-releases/knative.dev/eventing-contrib/cmd/event_display
EOF

kubectl apply -n test -f - <<EOF
apiVersion: "serving.kserve.io/v1beta1"
kind: "InferenceService"
metadata:
 name: "pmml-demo"
 namespace: "test"
spec:
 predictor:
   logger:
     mode: all
     url:  http://message-dumper.test
   pmml:
     storageUri:
  gs://kfserving-examples/models/pmml
EOF      
```

# Load Testing with iter8

```
$ iter8 k launch \
--set "tasks={ready,http,assess}" \
--set ready.isvc=pmml-demo \
--set ready.timeout=180s \
--set http.url=http://pmml-demo-predictor-default.test.svc.cluster.local/v1/models/pmml-demo:predict \
--set http.payloadURL=https://raw.githubusercontent.com/kserve/kserve/340df7a29d21d869a4f753db0dfa387c591635af/docs/samples/v1beta1/pmml/pmml-input.json \
--set http.contentType="application/json" \
--set assess.SLOs.upper.http/latency-mean=50 \
--set assess.SLOs.upper.http/error-count=0 \
--set runner=job

# Alternatively `hey` can be used with -m POST, -D <path_to_json> and -host arguments as in the curl command.

```
