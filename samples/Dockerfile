FROM  golang:1.20 AS builder
WORKDIR /go/src/github.com/opeshift/cluster-node-tuning-operator/samples/app
COPY . .
RUN mkdir build || true && go build -o build/sample-app ./app

FROM registry.access.redhat.com/ubi9/ubi
COPY --from=builder /go/src/github.com/opeshift/cluster-node-tuning-operator/samples/app/build/sample-app /bin/sample-app
RUN dnf install -y https://dl.fedoraproject.org/pub/epel/epel-release-latest-9.noarch.rpm && \
    dnf install -y htop procps && \
    dnf clean all
ENTRYPOINT ["/bin/sample-app"]
