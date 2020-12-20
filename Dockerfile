FROM golang:1.15-alpine AS builder

RUN apk update
RUN apk add --no-cache ca-certificates git

WORKDIR /go/src/github.com/e-travel/cloudwatchlogsbeat
COPY go.mod go.sum ./
RUN go mod download

COPY cwl cwl
COPY beater beater
COPY main.go .
RUN go mod vendor
RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -mod=vendor -i -o cloudwatchlogsbeat

FROM scratch

ARG BEAT_HOME="/usr/share/cloudwatchlogsbeat"

ENV PATH="${BEAT_HOME}:${PATH}"

COPY --from=builder /go/src/github.com/e-travel/cloudwatchlogsbeat/cloudwatchlogsbeat "${BEAT_HOME}/cloudwatchlogsbeat"
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
WORKDIR "${BEAT_HOME}"

CMD ["cloudwatchlogsbeat"]
