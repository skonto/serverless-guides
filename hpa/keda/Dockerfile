FROM registry.ci.openshift.org/openshift/release:golang-1.21 AS builder

ENV BASE=github.com/skonto/serverless-guides/hpa/keda
WORKDIR ${GOPATH}/src/${BASE}

COPY . .

ENV GOFLAGS="-mod=vendor"
RUN go build -o /tmp/echo ${BASE}/cmd/metrics

FROM openshift/origin-base
USER 65532

COPY --from=builder /tmp/echo /ko-app/metrics

