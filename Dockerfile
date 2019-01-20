FROM golang:1.11-alpine AS builder

WORKDIR /go/src/github.com/e-travel/cloudwatchlogsbeat
COPY . .
ENV CGO_ENABLED=0
RUN apk update && \
    apk add ca-certificates && \
    GOOS=linux GOARCH=amd64 go build -i -o cloudwatchlogsbeat


FROM scratch

ARG BEAT_HOME="/usr/share/cloudwatchlogsbeat"

ENV PATH="${BEAT_HOME}:${PATH}"

COPY --from=builder /go/src/github.com/e-travel/cloudwatchlogsbeat/cloudwatchlogsbeat "${BEAT_HOME}/cloudwatchlogsbeat"
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
WORKDIR "${BEAT_HOME}"

CMD ["cloudwatchlogsbeat"]