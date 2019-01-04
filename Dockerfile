FROM golang:1.11-stretch AS builder

WORKDIR /go/src/github.com/e-travel/cloudwatchlogsbeat
COPY . .
ENV CGO_ENABLED=0
RUN apt-get -y -qq update && \
    apt-get -y -qq install ca-certificates && \
    GOOS=linux GOARCH=amd64 go build -i -o cloudwatchlogsbeat


FROM debian:stretch

ARG DEBIAN_FRONTEND="noninteractive"
ARG BEAT_HOME="/usr/share/cloudwatchlogsbeat"

ENV PATH="${BEAT_HOME}:${PATH}"

RUN apt-get -y -qq update && apt-get -y -qq install ca-certificates
COPY --from=builder /go/src/github.com/e-travel/cloudwatchlogsbeat/cloudwatchlogsbeat "${BEAT_HOME}/cloudwatchlogsbeat"

WORKDIR "${BEAT_HOME}"

CMD ["/usr/bin/cloudwatchlogsbeat"]
