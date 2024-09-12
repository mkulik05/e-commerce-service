ARG ALPINE_VERSION=3.19.1
FROM alpine:${ALPINE_VERSION}

RUN --mount=type=cache,target=/var/cache/apk \
    apk --update add curl tar

ARG OS=linux
ARG ARCH=amd64
ARG MIGRATOR_VERSION=v4.17.1
RUN curl -L https://github.com/golang-migrate/migrate/releases/download/${MIGRATOR_VERSION}/migrate.${OS}-${ARCH}.tar.gz | tar xvz -C /bin/ && \
    ln -sf /bin/migrate /usr/bin/migrate

