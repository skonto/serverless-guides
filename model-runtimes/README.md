# Setup instructions

Adapted from 
https://github.com/pytorch/serve/blob/v0.11.0/examples/large_models/Huggingface_accelerate/Readme.md

```bash
Login into huggingface hub with token by running the below command

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

```
