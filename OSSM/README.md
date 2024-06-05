# Instructions

```
# install service mess and knative related resources
$ ./1_32/mesh.sh

# verify the setup
$ oc get pods -n istio-system
NAME                                   READY   STATUS    RESTARTS   AGE
istio-egressgateway-856fbd6dfc-t5xwv   1/1     Running   0          5m26s
istio-ingressgateway-7b6954747-qqhbw   1/1     Running   0          5m26s
istiod-basic-d6d4fdc9f-wczvv           1/1     Running   0          5m42s


$ oc get networkpolicy -n knative-serving
NAME                       POD-SELECTOR                   AGE
istio-expose-route-basic   maistra.io/expose-route=true   5m27s
istio-mesh-basic           <none>                         5m27s


# install serverless operator from OCP console or apply the next yaml and approve plan manually:
$ oc apply -f install_serverless.yaml

$ oc get installplans -n openshift-serverless
NAME            CSV                           APPROVAL    APPROVED
install-knqnh   serverless-operator.v1.32.1   Automatic   true

# verify installed Serverless operator pods
$ oc get pods -n openshift-serverless
NAME                                        READY   STATUS    RESTARTS   AGE
knative-openshift-545686887b-bx4wx          1/1     Running   0          43s
knative-openshift-ingress-97b46bb9-rv5js    1/1     Running   0          42s
knative-operator-webhook-5cf7789dbd-qmpnp   1/1     Running   0          44s

# install serving with istio enabled
$ oc apply -f serving.yaml

# verify that Knative Serving pods are up
$ oc get pods -n knative-serving
NAME                                                          READY   STATUS      RESTARTS   AGE
activator-7b646b88cc-fkrcw                                    2/2     Running     0          2m49s
activator-7b646b88cc-vntdm                                    2/2     Running     0          2m35s
autoscaler-588cfb64fb-8csmw                                   2/2     Running     0          2m49s
autoscaler-588cfb64fb-9nv45                                   2/2     Running     0          2m49s
autoscaler-hpa-684ccf4558-c6grm                               1/1     Running     0          2m48s
autoscaler-hpa-684ccf4558-nxztv                               1/1     Running     0          2m48s
controller-c85f6bb7-8wsmt                                     1/1     Running     0          2m43s
controller-c85f6bb7-wxvch                                     1/1     Running     0          2m48s
net-istio-controller-96d4b96dd-6bwc8                          1/1     Running     0          2m48s
net-istio-controller-96d4b96dd-r5tcz                          1/1     Running     0          2m48s
net-istio-webhook-5dfc9b8d7c-l8gx9                            1/1     Running     0          2m48s
net-istio-webhook-5dfc9b8d7c-vwl24                            1/1     Running     0          2m48s
storage-version-migration-serving-serving-1.11-1.32.1-r28pn   0/1     Completed   0          2m48s
webhook-5686f8c6f6-7q2cx                                      1/1     Running     0          2m34s
webhook-5686f8c6f6-rs9n5                                      1/1     Running     0          2m49s

# create a service
$ oc apply -f service.yaml -n user-ns

# verify that the service is up

$ oc get po -n user-ns
NAME                                              READY   STATUS    RESTARTS   AGE
helloworld-go-00001-deployment-5bd78dbc5f-xlp2f   3/3     Running   0          2m9s
helloworld-go-00001-deployment-5bd78dbc5f-zslnq   3/3     Running   0          2m9s

# verify the service
$ oc get ksvc -n user-ns
NAME            URL                                                                                        LATESTCREATED         LATESTREADY           READY   REASON
helloworld-go   https://helloworld-go-user-ns.apps.ci-ln-bxslict-76ef8.origin-ci-int-aws.dev.rhcloud.com   helloworld-go-00001   helloworld-go-00001   True  

$ $ curl -k   https://helloworld-go-user-ns.apps.ci-ln-bxslict-76ef8.origin-ci-int-aws.dev.rhcloud.com 
Hello Go Sample v1!

$ oc get ksvc -n user-ns -ojson | jq  '.items[].status.conditions'
[
  {
    "lastTransitionTime": "2024-06-05T10:59:19Z",
    "status": "True",
    "type": "ConfigurationsReady"
  },
  {
    "lastTransitionTime": "2024-06-05T10:59:20Z",
    "status": "True",
    "type": "Ready"
  },
  {
    "lastTransitionTime": "2024-06-05T10:59:20Z",
    "status": "True",
    "type": "RoutesReady"
  }
]


$ oc get  networkpolicy  -n user-ns
NAME                       POD-SELECTOR                   AGE
istio-expose-route-basic   maistra.io/expose-route=true   103s
istio-mesh-basic           <none>                         103s


$ oc get routes.serving.knative.dev -n user-ns
NAME            URL                                                                                        READY   REASON
helloworld-go   https://helloworld-go-user-ns.apps.ci-ln-4zhdf4b-72292.origin-ci-int-gce.dev.rhcloud.com   True    

$ oc get ingresses.networking.internal.knative.dev -n user-ns
NAME            READY   REASON
helloworld-go   True    

$  oc get route -n istio-system
NAME                                                      HOST/PORT                                                                                      PATH   SERVICES               PORT    TERMINATION            WILDCARD
istio-ingressgateway                                      istio-ingressgateway-istio-system.apps.ci-ln-bxslict-76ef8.origin-ci-int-aws.dev.rhcloud.com          istio-ingressgateway   8080                           None
route-acbd3525-dae1-4f5f-b979-0093b802b28b-313336626532   helloworld-go-user-ns.apps.ci-ln-bxslict-76ef8.origin-ci-int-aws.dev.rhcloud.com                      istio-ingressgateway   https   passthrough/Redirect   None

$ oc get revision -n user-ns
NAME                  CONFIG NAME     K8S SERVICE NAME   GENERATION   READY   REASON   ACTUAL REPLICAS   DESIRED REPLICAS
helloworld-go-00001   helloworld-go                      1            True             1                 1

$ oc get configuration -n user-ns
NAME            LATESTCREATED         LATESTREADY           READY   REASON
helloworld-go   helloworld-go-00001   helloworld-go-00001   True    

$ oc get service -n user-ns
NAME                          TYPE           CLUSTER-IP      EXTERNAL-IP                                            PORT(S)                                              AGE
helloworld-go                 ExternalName   <none>          knative-local-gateway.istio-system.svc.cluster.local   80/TCP                                               3m
helloworld-go-00001           ClusterIP      172.30.61.176   <none>                                                 80/TCP,443/TCP                                       3m6s
helloworld-go-00001-private   ClusterIP      172.30.201.0    <none>       

$ oc get virtualservices.networking.istio.io -n user-ns
NAME                    GATEWAYS                                                                              HOSTS                                                                                                                                                                                AGE
helloworld-go-ingress   ["knative-serving/knative-ingress-gateway","knative-serving/knative-local-gateway"]   ["helloworld-go-user-ns.apps.ci-ln-bxslict-76ef8.origin-ci-int-aws.dev.rhcloud.com","helloworld-go.user-ns","helloworld-go.user-ns.svc","helloworld-go.user-ns.svc.cluster.local"]   3m35s
helloworld-go-mesh      ["mesh"]                                                                              ["helloworld-go.user-ns","helloworld-go.user-ns.svc","helloworld-go.user-ns.svc.cluster.local"]                                                                                      3m35s

```

To get must-gather for serverless:

```
 oc adm must-gather --image=registry.redhat.io/openshift-serverless-1/svls-must-gather-rhel8:1.32.0
```

To get must-gather for OSSM:

```
oc adm must-gather --image=registry.redhat.io/openshift-service-mesh/istio-must-gather-rhel8:2.4
```
