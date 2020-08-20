FROM alpine:3.11 AS downloader

RUN apk add --no-cache \
    ca-certificates \
    libc6-compat \
    libffi-dev \
    linux-headers \
    netcat-openbsd \
    tzdata 

COPY dist/worldr-push-proxy.tar.gz /tmp/worldr-push-proxy.tar.gz

RUN set -ex \
    && tar -xvzf /tmp/worldr-push-proxy.tar.gz --directory / \
    && chown -R nobody:nogroup /worldr-push-proxy

USER nobody

WORKDIR /worldr-push-proxy

ENV PUSH_PROXY=/worldr-push-proxy/bin/worldr-push-proxy

COPY docker/entrypoint.sh /usr/local/bin/

EXPOSE 8066

VOLUME ["/worldr-push-proxy/config", "/worldr-push-proxy/certs"]

ENTRYPOINT ["/usr/local/bin/entrypoint.sh"]
