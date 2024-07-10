ARG BASE_IMAGE=pytorch/torchserve:0.11.0-gpu

FROM ${BASE_IMAGE}

COPY custom_handler.py \
     config.properties \
     requirements.txt \
     model.zip \
     setup_config.json /home/model-server/

RUN torch-model-archiver --model-name flan-t5-small --version 1.0 --handler custom_handler.py --extra-files model.zip,setup_config.json -r requirements.txt

FROM ${BASE_IMAGE}

ENV PATH=/home/model-server/.local/bin:$PATH
ENV TS_CONFIG_FILE=/home/model-server/config.properties
ENV TF_CPP_MIN_LOG_LEVEL=2

COPY --from=0 /home/model-server/flan-t5-small.mar /home/model-server/model-store
COPY config.properties /home/model-server/
