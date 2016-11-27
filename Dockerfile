FROM alpine:latest

COPY bin/linkerd-proxy /usr/local/bin/linkerd-proxy

CMD linkerd-proxy
