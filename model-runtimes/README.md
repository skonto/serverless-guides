# Setup instructions

Adapted from 
https://github.com/pytorch/serve/blob/v0.11.0/examples/large_models/Huggingface_accelerate/Readme.md

```bash
#Login into huggingface hub with token by running the below command

huggingface-cli login

#paste the token generated from huggingface hub.

# use the script https://github.com/pytorch/serve/blob/v0.11.0/examples/large_models/Huggingface_accelerate/Download_model.py
python Download_model.py --model_name flan-t5-small --model_path ./model

cd model/models--google--flan-t5-small/
zip -r ../../model.zip *
cd ../..
docker build -t docker.io/skonto/flan-t5-small:gpu .

# this is slow it takes ~15min locally as it installs the python deps and loads the model in cpu memory (haven't experimented with gpu)
# no optimization has been done, meant to explore the metrics api for autoscaling 
docker run -p 8080:8080 -p 8082:8082 -it docker.io/skonto/flan-t5-small:gpu

curl "http://localhost:8080/predictions/flan-t5-small"
curl localhost:8082/metrics
curl "http://localhost:8080/ping"


On a real cluster the flan model comes up much faster ~3min:

oc apply -f flan-small.yaml

oc get ksvc
NAME         URL                                                                          LATESTCREATED      LATESTREADY        READY   REASON
flan-small   https://flan-small-default.apps.ci-ln-2ipm0m2-76ef8.aws-2.ci.openshift.org   flan-small-00001   flan-small-00001   True 

oc get po
NAME                                           READY   STATUS    RESTARTS   AGE
flan-small-00001-deployment-6fbc676dc6-tcd7h   2/2     Running   0          14m

kubectl top pod flan-small-00001-deployment-6fbc676dc6-tcd7h 
NAME                                           CPU(cores)   MEMORY(bytes)
flan-small-00001-deployment-6fbc676dc6-tcd7h   3m           4712Mi

curl  -k https://flan-small-default..../predictions/flan-t5-small -T ./sample_txt.txt
a sunny day with the wind and lots of clouds

oc port-forward flan-small-00001-deployment-6fbc676dc6-tcd7h 8082:8082
Forwarding from 127.0.0.1:8082 -> 8082
Forwarding from [::1]:8082 -> 8082
Handling connection for 8082


curl localhost:8082/metrics
# HELP DiskUtilization Torchserve prometheus gauge metric with unit: Percent
# TYPE DiskUtilization gauge
DiskUtilization{Level="Host",Hostname="flan-small-00001-deployment-6fbc676dc6-tcd7h",} 32.0
# HELP DiskAvailable Torchserve prometheus gauge metric with unit: Gigabytes
# TYPE DiskAvailable gauge
DiskAvailable{Level="Host",Hostname="flan-small-00001-deployment-6fbc676dc6-tcd7h",} 81.23770141601562
# HELP GPUMemoryUsed Torchserve prometheus gauge metric with unit: Megabytes
# TYPE GPUMemoryUsed gauge
# HELP WorkerThreadTime Torchserve prometheus gauge metric with unit: Milliseconds
# TYPE WorkerThreadTime gauge
WorkerThreadTime{Level="Host",Hostname="flan-small-00001-deployment-6fbc676dc6-tcd7h",} 2.0
# HELP HandlerTime Torchserve prometheus gauge metric with unit: ms
# TYPE HandlerTime gauge
HandlerTime{ModelName="flan-t5-small",Level="Model",Hostname="flan-small-00001-deployment-6fbc676dc6-tcd7h",} 620.28
# HELP GPUUtilization Torchserve prometheus gauge metric with unit: Percent
# TYPE GPUUtilization gauge
# HELP GPUMemoryUtilization Torchserve prometheus gauge metric with unit: Percent
# TYPE GPUMemoryUtilization gauge
# HELP ts_queue_latency_microseconds Torchserve prometheus counter metric with unit: Microseconds
# TYPE ts_queue_latency_microseconds counter
ts_queue_latency_microseconds{model_name="flan-t5-small",model_version="default",hostname="flan-small-00001-deployment-6fbc676dc6-tcd7h",} 461.933
# HELP PredictionTime Torchserve prometheus gauge metric with unit: ms
# TYPE PredictionTime gauge
PredictionTime{ModelName="flan-t5-small",Level="Model",Hostname="flan-small-00001-deployment-6fbc676dc6-tcd7h",} 620.43
# HELP Requests4XX Torchserve prometheus counter metric with unit: Count
# TYPE Requests4XX counter
Requests4XX{Level="Host",Hostname="flan-small-00001-deployment-6fbc676dc6-tcd7h",} 2.0
# HELP ts_inference_requests_total Torchserve prometheus counter metric with unit: Count
# TYPE ts_inference_requests_total counter
ts_inference_requests_total{model_name="flan-t5-small",model_version="default",hostname="flan-small-00001-deployment-6fbc676dc6-tcd7h",} 4.0
# HELP MemoryAvailable Torchserve prometheus gauge metric with unit: Megabytes
# TYPE MemoryAvailable gauge
MemoryAvailable{Level="Host",Hostname="flan-small-00001-deployment-6fbc676dc6-tcd7h",} 11292.19140625
# HELP ts_inference_latency_microseconds Torchserve prometheus counter metric with unit: Microseconds
# TYPE ts_inference_latency_microseconds counter
ts_inference_latency_microseconds{model_name="flan-t5-small",model_version="default",hostname="flan-small-00001-deployment-6fbc676dc6-tcd7h",} 2511000.058
# HELP QueueTime Torchserve prometheus gauge metric with unit: Milliseconds
# TYPE QueueTime gauge
QueueTime{Level="Host",Hostname="flan-small-00001-deployment-6fbc676dc6-tcd7h",} 0.0
# HELP MemoryUsed Torchserve prometheus gauge metric with unit: Megabytes
# TYPE MemoryUsed gauge
MemoryUsed{Level="Host",Hostname="flan-small-00001-deployment-6fbc676dc6-tcd7h",} 3911.49609375
# HELP Requests2XX Torchserve prometheus counter metric with unit: Count
# TYPE Requests2XX counter
Requests2XX{Level="Host",Hostname="flan-small-00001-deployment-6fbc676dc6-tcd7h",} 79.0
# HELP DiskUsage Torchserve prometheus gauge metric with unit: Gigabytes
# TYPE DiskUsage gauge
DiskUsage{Level="Host",Hostname="flan-small-00001-deployment-6fbc676dc6-tcd7h",} 38.19880294799805
# HELP Requests5XX Torchserve prometheus counter metric with unit: Count
# TYPE Requests5XX counter
# HELP MemoryUtilization Torchserve prometheus gauge metric with unit: Percent
# TYPE MemoryUtilization gauge
MemoryUtilization{Level="Host",Hostname="flan-small-00001-deployment-6fbc676dc6-tcd7h",} 27.8
# HELP CPUUtilization Torchserve prometheus gauge metric with unit: Percent
# TYPE CPUUtilization gauge
CPUUtilization{Level="Host",Hostname="flan-small-00001-deployment-6fbc676dc6-tcd7h",} 0.0
# HELP WorkerLoadTime Torchserve prometheus gauge metric with unit: Milliseconds
# TYPE WorkerLoadTime gauge
WorkerLoadTime{WorkerName="W-9000-flan-t5-small_1.0",Level="Host",Hostname="flan-small-00001-deployment-6fbc676dc6-tcd7h",} 87801.0
WorkerLoadTime{WorkerName="W-9002-flan-t5-small_1.0",Level="Host",Hostname="flan-small-00001-deployment-6fbc676dc6-tcd7h",} 194075.0
WorkerLoadTime{WorkerName="W-9003-flan-t5-small_1.0",Level="Host",Hostname="flan-small-00001-deployment-6fbc676dc6-tcd7h",} 142320.0
WorkerLoadTime{WorkerName="W-9001-flan-t5-small_1.0",Level="Host",Hostname="flan-small-00001-deployment-6fbc676dc6-tcd7h",} 223259.0


```

