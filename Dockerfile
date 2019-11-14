FROM golang:1.11 as builder

ENV GO111MODULE=on
ENV GOPROXY=https://goproxy.io

COPY . /go/src/github.com/fusion-app/ResourceProbeExample

RUN cd /go/src/github.com/fusion-app/ResourceProbeExample && go build -v

FROM registry.njuics.cn/library/ubuntu:18.04

COPY --from=builder /go/src/github.com/fusion-app/ResourceProbeExample/ResourceProbeExample /root/prober

ENTRYPOINT ["/root/prober"]
