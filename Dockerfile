FROM golang:1.13-alpine3.10 as builder

WORKDIR /src
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags '-extldflags "-static"' -mod vendor -o openstackswift-exporter .

FROM alpine:3.10
COPY --from=builder /src/openstackswift-exporter /opt/openstackswift-exporter

CMD [ \
    "/opt/openstackswift-exporter", \
    "-swift-user-name=swift-user", \
    "-swift-password=pass", \
    "-swift-use-insecure-tls" \
     ]
