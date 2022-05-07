FROM golang:1.18-alpine as build
ARG BUILD_VERSION=docker
ARG TARGETOS
ARG TARGETARCH
ENV GOOS=$TARGETOS
ENV GOARCH=$TARGETARCH
ENV GOPROXY https://proxy.golang.org,direct
ENV GOSUMDB off
RUN apk add -q tzdata make unzip
WORKDIR /src
COPY . .
RUN make build BUILD_VERSION=$BUILD_VERSION
RUN mv bin/${GOOS}-${GOARCH} binaries
RUN mkdir tmp_dir

FROM scratch
WORKDIR /
COPY --from=build /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=build /src/tmp_dir /tmp
COPY --from=build /src/binaries bin
COPY configs configs
COPY web web
ENTRYPOINT ["server"]
