FROM alpine:3.11 AS downloader

RUN apk add --no-cache \
    ca-certificates \
    libc6-compat \
    libffi-dev \
    linux-headers \
    netcat-openbsd \
    tzdata 

COPY . /worldr-push-proxy

RUN set -ex && chown -R nobody:nogroup /worldr-push-proxy

USER nobody

WORKDIR /worldr-push-proxy

ENV PUSH_PROXY=/worldr-push-proxy/bin/worldr-push-proxy

COPY docker/entrypoint.sh /usr/local/bin/

EXPOSE 8066

VOLUME ["/worldr-push-proxy/config", "/worldr-push-proxy/certs"]

ENTRYPOINT ["/usr/local/bin/entrypoint.sh"]
