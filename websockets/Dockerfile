FROM registry.ci.openshift.org/openshift/release:golang-1.22 AS builder

ENV BASE=github.com/skonto/serverless-guides/websockets
WORKDIR ${GOPATH}/src/${BASE}

COPY . .

ENV GOFLAGS="-mod=vendor"
RUN go build -o /tmp/echo ${BASE}/cmd/ws

FROM openshift/origin-base
USER 65532

COPY --from=builder /tmp/echo /ko-app/ws
