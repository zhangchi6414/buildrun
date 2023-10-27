FROM golang:1.19.8  as builder

WORKDIR /go/src/buildRun/
COPY cmd/ cmd/
COPY pkg/ pkg/
COPY utils/ utils/


# Build
RUN  go env -w GOPROXY=https://goproxy.cn,direct && \
     cd /go/src/buildRun/ && \
     go mod init buildRun && \
     go mod tidy && \
     CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o builder /go/src/buildRun/cmd/main.go

FROM 192.168.2.106:1180/fangzhou/alpine:v1.0.0

WORKDIR /root/

RUN mkdir /root/docker/
#RUN apk update && apk upgrade && \
#    apk add --no-cache bash git openssh

COPY --from=builder /go/src/buildRun/builder .
CMD ["./builder"]

