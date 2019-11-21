FROM golang:1.11 as builder

ENV GO111MODULE=on
ENV GOPROXY=https://goproxy.io

COPY . /go/src/github.com/fusion-app/prober

RUN cd /go/src/github.com/fusion-app/prober && go build -v github.com/fusion-app/prober/cmd/instance-prober

FROM registry.njuics.cn/library/ubuntu:18.04

COPY --from=builder /go/src/github.com/fusion-app/prober/instance-prober /root/prober

ENTRYPOINT ["/root/prober"]
