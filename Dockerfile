FROM golang:1.11.0-alpine3.8 AS builder
WORKDIR /go/src/github.com/e-travel/cloudwatchlogsbeat
COPY . .
ENV CGO_ENABLED=0
RUN apk update && \
    apk add -U ca-certificates && \
    GOOS=linux GOARCH=amd64 go build -i -o cloudwatchlogsbeat

FROM scratch
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /go/src/github.com/e-travel/cloudwatchlogsbeat/cloudwatchlogsbeat /cloudwatchlogsbeat
CMD ["/cloudwatchlogsbeat"]
