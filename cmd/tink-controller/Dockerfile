FROM alpine:3.11
ENTRYPOINT ["/usr/bin/tink-controller"]
EXPOSE 42113
EXPOSE 42114

ARG TARGETARCH
ARG TARGETVARIANT

RUN apk add --update ca-certificates && \
    apk add --repository=http://dl-cdn.alpinelinux.org/alpine/edge/testing cfssl
COPY tink-controller-linux-${TARGETARCH:-amd64}${TARGETVARIANT} /usr/bin/tink-controller
