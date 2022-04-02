FROM golang:1.18 as debug
WORKDIR /src
ENTRYPOINT ["make"]

FROM debug as build
ARG BUILD_VERSION=docker
ARG TARGETOS
ARG TARGETARCH
ENV GOOS=$TARGETOS
ENV GOARCH=$TARGETARCH
ENV GOPROXY https://proxy.golang.org,direct
ENV GOSUMDB off
COPY . .
RUN make build BUILD_VERSION=$BUILD_VERSION

FROM scratch
WORKDIR /app
COPY --from=build /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=build /src/bin bin
COPY configs configs
COPY web web
VOLUME ["/app/var", "/app/web/library"]
ENTRYPOINT ["bin/server"]
