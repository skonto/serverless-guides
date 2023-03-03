# Instructions

```
# install service mess and knative related resources
$ ./mesh.sh

# install serverless operator from OCP console or apply the next yaml and approve plan manually:
$ oc apply -f install_serverless.yaml

# install serving with istio enabled
$ oc apply -f serving.yaml

# create a service
$ oc apply -f service.yaml

# verify the service
$ oc get ksvc -n user-ns
NAME            URL                                                                             LATESTCREATED         LATESTREADY           READY   REASON
helloworld-go   https://helloworld-go-user-ns.apps.ci-ln-hbfib0t-72292.gcp-2.ci.openshift.org   helloworld-go-00001   helloworld-go-00001   True    

$ curl -k https://helloworld-go-user-ns.apps.ci-ln-hbfib0t-72292.gcp-2.ci.openshift.org
Hello Go Sample v1!
```
