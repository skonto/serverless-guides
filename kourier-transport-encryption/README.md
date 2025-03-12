# Kourier transport encryption
 
 ```
 # Install cert-manager and wait for the installation to be sucessful
 oc apply -f certmanager_subscription.yaml 
 
 # Apply the cet-manager resources
 oc apply -f serving-selfsigned-issuer.yaml
 oc apply -f serving-ca-issuer.yaml
 oc apply -n cert-manager -f serving-ca-certificate.yaml

oc get secret -n cert-manager "knative-selfsigned-ca" -o=jsonpath='{.data.tls\.crt}' | base64 -d > tls.crt
oc get secret -n cert-manager "knative-selfsigned-ca" -o=jsonpath='{.data.ca\.crt}' | base64 -d > ca.crt

# craete/sync trust bundles
for ns in "knative-serving" "knative-serving-ingress" "test"; do
   echo "Syncing trust-bundle for namespace: ${ns}"
   oc create namespace "${ns}" --dry-run=client -o yaml | oc apply -f -
   oc label namespace "${ns}" knative.openshift.io/part-of="openshift-serverless" --overwrite
   oc create configmap -n "${ns}" knative-ca-bundle --from-file=tls.crt --from-file=ca.crt \
        --dry-run=client -o yaml | kubectl apply -n "${ns}" -f -
   oc label configmap -n "${ns}" knative-ca-bundle networking.knative.dev/trust-bundle=true --overwrite
done

rm -f tls.crt ca.crt

# Install Serverless Operator, Serving and wait for the installation to be sucessful

oc apply -f serverless_subscription.yaml 

cat <<EOF | oc apply -n knative-serving -f -
apiVersion: operator.knative.dev/v1beta1
kind: KnativeServing
metadata:
  name: knative-serving
spec:
  config:
    certmanager:
      clusterLocalIssuerRef: |
        kind: ClusterIssuer
        name: knative-selfsigned-issuer
      systemInternalIssuerRef: |
        kind: ClusterIssuer
        name: knative-selfsigned-issuer
    network:
      system-internal-tls: "enabled"
      cluster-local-domain-tls: "enabled"
      httpProtocol: redirected
      openshift-ingress-default-certificate: wildcard
EOF

# Note:  if you already have Serving running and enable encryption you need to restart the following pods
# Restart controller to enable cert-manager integration
oc delete pod -n knative-serving -l app=controller
oc wait --timeout=60s --for=condition=Available deployment  -n knative-serving controller

# Restart activator to mount the certificates
oc delete pod -n knative-serving -l app=activator
oc wait --timeout=60s --for=condition=Available deployment  -n knative-serving activator

# apply the app in the test ns

oc get route -n knative-serving-ingress
NAME                                                      HOST/PORT                                                            PATH   SERVICES   PORT    TERMINATION            WILDCARD
route-e05bde07-1f60-479f-9825-80cd3187a64b-613164383836   helloworld-go-test.apps.ci-ln-y9mfc8b-76ef8.aws-2.ci.openshift.org          kourier    https   passthrough/Redirect   None

curl -k  https://helloworld-go-test.apps.ci-ln-y9mfc8b-76ef8.aws-2.ci.openshift.org
Hello Go Sample v1!
```


## Using a non default OCP Ingress certificate

See https://docs.openshift.com/container-platform/4.18/security/certificates/replacing-default-ingress-certificate.html for more,

```
out_dir=./
openssl req -x509 -sha256 -nodes -days 365 -newkey rsa:2048 \
  -subj "/O=Example Inc./CN=Example" \
  -keyout "${out_dir}"/root.key \
  -out "${out_dir}"/root.crt

subdomain=$(oc get ingresses.config.openshift.io cluster -o jsonpath="{.spec.domain}")

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

oc create configmap custom-ca \
     --from-file=ca-bundle.crt="${out_dir}"/root.crt \
     -n openshift-config

oc patch proxy/cluster \
     --type=merge \
     --patch='{"spec":{"trustedCA":{"name":"custom-ca"}}}'

oc create secret tls wildcard \
     --cert="${out_dir}"/wildcard.crt \
     --key="${out_dir}"/wildcard.key  \
     -n openshift-ingress
     
oc patch ingresscontroller.operator default \
    --type=merge -p \
    '{"spec":{"defaultCertificate": {"name": "wildcard"}}}' \
    -n openshift-ingress-operator
    

# Apply the Serving CR with the right OCP cert
    
cat <<EOF | oc apply -n knative-serving -f -
apiVersion: operator.knative.dev/v1beta1
kind: KnativeServing
metadata:
  name: knative-serving
spec:
  config:
    certmanager:
      clusterLocalIssuerRef: |
        kind: ClusterIssuer
        name: knative-selfsigned-issuer
      systemInternalIssuerRef: |
        kind: ClusterIssuer
        name: knative-selfsigned-issuer
    network:
      system-internal-tls: "enabled"
      cluster-local-domain-tls: "enabled"
      httpProtocol: redirected
      openshift-ingress-default-certificate: wildcard
EOF

# Check that the right env var is set when the net-kourier pods are restarted

oc get po -lapp=net-kourier-controller -n knative-serving-ingress -o json | jq -r '.items[0].spec.containers[] | select(.env) | .env[] | select(.name=="CERTS_SECRET_NAMESPACE") | .value'
openshift-ingress

oc get po -lapp=net-kourier-controller -n knative-serving-ingress -o json | jq -r '.items[0].spec.containers[] | select(.env) | .env[] | select(.name=="CERTS_SECRET_NAME") | .value'
wildcard

# Delete and then apply the app in the test ns

curl -k  https://helloworld-go-test.apps.ci-ln-y9mfc8b-76ef8.aws-2.ci.openshift.org
Hello Go Sample v1!
```

