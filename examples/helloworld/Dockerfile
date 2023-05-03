FROM golang:1.20-alpine3.17 as builder

ADD . /workspace
WORKDIR /workspace
RUN go mod download && go build -a -ldflags '-s' -o main

FROM alpine:3.17

RUN apk add ca-certificates

LABEL cqless-gateway="latest"

COPY --from=builder /workspace/main /usr/bin/server

WORKDIR /workspace

ENTRYPOINT ["server"]