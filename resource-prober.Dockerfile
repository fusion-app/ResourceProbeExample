FROM golang:1.14-alpine3.12 as builder

ENV GOPATH /go
ENV GOPROXY=https://goproxy.cn

COPY . /go/src/github.com/fusion-app/prober

RUN go build -o /resource-prober /go/src/github.com/fusion-app/prober/cmd/resource-prober

FROM alpine:3.10

COPY --from=builder /resource-prober /root/resource-prober

ENTRYPOINT ["/root/resource-prober"]
