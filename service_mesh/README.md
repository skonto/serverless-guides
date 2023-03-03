# Instructions

```
# install service mess and knative related resources
$ ./mesh.sh

# verify the setup
$ oc get pods -n istio-system
NAME                                   READY   STATUS    RESTARTS   AGE
grafana-5bf9788595-56tp4               2/2     Running   0          32m
istio-egressgateway-757d5fb65d-d2hcz   1/1     Running   0          32m
istio-ingressgateway-9d96c6789-mvjmw   1/1     Running   0          32m
istiod-basic-544d455f65-7qpmd          1/1     Running   0          33m
jaeger-7b7c8dc85b-fcncz                2/2     Running   0          32m
kiali-74dd54dd5d-xz8hl                 1/1     Running   0          30m
prometheus-ff79665c9-scskr             2/2     Running   0          32m
wasm-cacher-basic-6f5c57745c-mqr7g     1/1     Running   0          32m

$ oc get networkpolicy -n knative-serving
NAME                                 POD-SELECTOR                   AGE
allow-from-openshift-monitoring-ns   <none>                         26m
domainmapping-webhook                app=domainmapping-webhook      26m
istio-expose-route-basic             maistra.io/expose-route=true   32m
istio-mesh-basic                     <none>                         32m
net-istio-webhook                    app=net-istio-webhook          26m
webhook                              app=webhook                    26m

# install serverless operator from OCP console or apply the next yaml and approve plan manually:
$ oc apply -f install_serverless.yaml

$ oc get installplans -n openshift-serverless
NAME            CSV                           APPROVAL   APPROVED
install-r6btg   serverless-operator.v1.26.0   Manual     false

# patch the install plan
$  oc -n openshift-serverless patch installplan install-r6btg -p '{"spec":{"approved":true}}' --type merge
installplan.operators.coreos.com/install-r6btg patched

# verify installed Serverless operator pods
$ oc get pods -n openshift-serverless
NAME                                         READY   STATUS    RESTARTS   AGE
knative-openshift-78f7d944d7-z5dm6           1/1     Running   0          72s
knative-openshift-ingress-69dd8bc4cb-8djmp   1/1     Running   0          69s
knative-operator-webhook-598454bc99-gdhwf    1/1     Running   0          73s

# install serving with istio enabled
$ oc apply -f serving.yaml

# verify that Knative Serving pods are up
$ oc get pods -n knative-serving
NAME                                                           READY   STATUS      RESTARTS   AGE
activator-fc9fcf778-9dxbs                                      2/2     Running     0          56s
activator-fc9fcf778-dl9r2                                      2/2     Running     0          56s
autoscaler-67df785744-m8k8c                                    2/2     Running     0          56s
autoscaler-67df785744-x2ph9                                    2/2     Running     0          56s
autoscaler-hpa-d45bfc87b-fwtwq                                 1/1     Running     0          54s
autoscaler-hpa-d45bfc87b-wd9bd                                 1/1     Running     0          54s
controller-65bcd69f-9hs8n                                      1/1     Running     0          53s
controller-65bcd69f-sfjlw                                      1/1     Running     0          46s
domain-mapping-579474f678-rl2nb                                1/1     Running     0          55s
domain-mapping-579474f678-vl9bc                                1/1     Running     0          55s
domainmapping-webhook-79986f7545-mrbqk                         1/1     Running     0          55s
domainmapping-webhook-79986f7545-qkpcv                         1/1     Running     0          55s
net-istio-controller-65dd6945cc-h59zp                          1/1     Running     0          53s
net-istio-controller-65dd6945cc-rsx42                          1/1     Running     0          53s
net-istio-webhook-5784f45744-7rnhq                             1/1     Running     0          52s
net-istio-webhook-5784f45744-lt4zd                             1/1     Running     0          53s
storage-version-migration-serving-serving-1.5.0-1.26.0-rxdwg   0/1     Completed   0          54s
webhook-778d5987b7-dvxnm                                       1/1     Running     0          55s
webhook-778d5987b7-q7nvk                                       1/1     Running     0          54s


# create a service
$ oc apply -f service.yaml -n user-ns

# verify that the service is up

$ oc get pods -n user-ns
NAME                                              READY   STATUS    RESTARTS   AGE
helloworld-go-00001-deployment-7f6f845c67-cs9pv   3/3     Running   0          21s


# verify the service
$ oc get ksvc -n user-ns
NAME            URL                                                                                        LATESTCREATED         LATESTREADY           READY   REASON
helloworld-go   https://helloworld-go-user-ns.apps.ci-ln-4zhdf4b-72292.origin-ci-int-gce.dev.rhcloud.com   helloworld-go-00001   helloworld-go-00001   True   

$ curl -k https://helloworld-go-user-ns.apps.ci-ln-4zhdf4b-72292.origin-ci-int-gce.dev.rhcloud.com
Hello Go Sample v1!

$ oc get ksvc -n user-ns -ojson | jq  '.items[].status.conditions'
[
  {
    "lastTransitionTime": "2023-03-03T20:12:09Z",
    "status": "True",
    "type": "ConfigurationsReady"
  },
  {
    "lastTransitionTime": "2023-03-03T20:12:09Z",
    "status": "True",
    "type": "Ready"
  },
  {
    "lastTransitionTime": "2023-03-03T20:12:09Z",
    "status": "True",
    "type": "RoutesReady"
  }
]

$ oc get  networkpolicy  -n user-ns
NAME                       POD-SELECTOR                   AGE
istio-expose-route-basic   maistra.io/expose-route=true   14m
istio-mesh-basic           <none>                         14m

$ oc get routes.serving.knative.dev -n user-ns
NAME            URL                                                                                        READY   REASON
helloworld-go   https://helloworld-go-user-ns.apps.ci-ln-4zhdf4b-72292.origin-ci-int-gce.dev.rhcloud.com   True    

$ oc get ingresses.networking.internal.knative.dev -n user-ns
NAME            READY   REASON
helloworld-go   True    


$ oc get route -n istio-system
NAME                                                      HOST/PORT                                                                                      PATH   SERVICES               PORT          TERMINATION            WILDCARD
grafana                                                   grafana-istio-system.apps.ci-ln-4zhdf4b-72292.origin-ci-int-gce.dev.rhcloud.com                       grafana                <all>         reencrypt/Redirect     None
istio-ingressgateway                                      istio-ingressgateway-istio-system.apps.ci-ln-4zhdf4b-72292.origin-ci-int-gce.dev.rhcloud.com          istio-ingressgateway   8080                                 None
jaeger                                                    jaeger-istio-system.apps.ci-ln-4zhdf4b-72292.origin-ci-int-gce.dev.rhcloud.com                        jaeger-query           https-query   reencrypt              None
kiali                                                     kiali-istio-system.apps.ci-ln-4zhdf4b-72292.origin-ci-int-gce.dev.rhcloud.com                         kiali                  20001         reencrypt/Redirect     None
prometheus                                                prometheus-istio-system.apps.ci-ln-4zhdf4b-72292.origin-ci-int-gce.dev.rhcloud.com                    prometheus             <all>         reencrypt/Redirect     None
route-d2ba8638-1c3e-4aad-a540-4cadff8e1f9a-376436376534   helloworld-go-user-ns.apps.ci-ln-4zhdf4b-72292.origin-ci-int-gce.dev.rhcloud.com                      istio-ingressgateway   https         passthrough/Redirect   None

$ oc get revision -n user-ns
NAME                  CONFIG NAME     K8S SERVICE NAME   GENERATION   READY   REASON   ACTUAL REPLICAS   DESIRED REPLICAS
helloworld-go-00001   helloworld-go                      1            True             1                 1

$ oc get configuration -n user-ns
NAME            LATESTCREATED         LATESTREADY           READY   REASON
helloworld-go   helloworld-go-00001   helloworld-go-00001   True    

$ oc get service -n user-ns
NAME                          TYPE           CLUSTER-IP       EXTERNAL-IP                                            PORT(S)                                              AGE
helloworld-go                 ExternalName   <none>           knative-local-gateway.istio-system.svc.cluster.local   80/TCP                                               21m
helloworld-go-00001           ClusterIP      172.30.135.211   <none>                                                 80/TCP,443/TCP                                       21m
helloworld-go-00001-private   ClusterIP      172.30.23.80     <none>     
```

To get must-gather for serverless:

```
oc adm must-gather --image=quay.io/openshift-knative/must-gather
```

To get must-gather for istio:

```
oc adm must-gather --image=registry.redhat.io/openshift-service-mesh/istio-must-gather-rhel8:2.1.1-1
```
