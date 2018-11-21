FROM golang:1.11-alpine as builder

WORKDIR /go/src/github.com/kmacoskey/taos

COPY . .

RUN apk --no-cache add curl git make && \
    go get ./... && \
    make build && \
    curl -LO https://releases.hashicorp.com/terraform/0.11.10/terraform_0.11.10_linux_amd64.zip && \
    unzip terraform_0.11.10_linux_amd64.zip terraform -d /usr/local/bin/ && \
    rm terraform_0.11.10_linux_amd64.zip

FROM alpine:latest
RUN apk --no-cache add ca-certificates

COPY --from=builder /go/src/github.com/kmacoskey/taos/taos .
COPY --from=builder /usr/local/bin/terraform /usr/local/bin/terraform

CMD ["./taos"]
