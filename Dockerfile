FROM golang:1.18 as build
ARG BUILD_VERSION=docker
ARG TARGETOS
ARG TARGETARCH
ENV GOOS=$TARGETOS
ENV GOARCH=$TARGETARCH
ENV GOPROXY https://proxy.golang.org,direct
ENV GOSUMDB off
WORKDIR /src
COPY . .
RUN make build-indexer BUILD_VERSION=$BUILD_VERSION
RUN make build-server BUILD_VERSION=$BUILD_VERSION
RUN make build-converter
RUN mkdir tmp_dir

FROM scratch
WORKDIR /
COPY --from=build /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=build /src/tmp_dir /tmp
COPY --from=build /src/bin bin
COPY configs configs
COPY web web
VOLUME ["/var/index"]
ENTRYPOINT ["server"]
