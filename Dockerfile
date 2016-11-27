FROM alpine:latest

RUN apk --no-cache add ca-certificates

COPY bin/linkerd-proxy /usr/local/bin/linkerd-proxy

CMD linkerd-proxy
