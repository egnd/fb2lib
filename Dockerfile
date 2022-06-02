FROM golang:1.18-alpine as build
ARG BUILD_VERSION=docker
ARG TARGETOS
ARG TARGETARCH
ENV GOOS=$TARGETOS
ENV GOARCH=$TARGETARCH
ENV GOPROXY https://proxy.golang.org,direct
ENV GOSUMDB off
RUN apk add -q tzdata make unzip curl ca-certificates
WORKDIR /src
COPY . .
RUN make build BUILD_VERSION=$BUILD_VERSION
RUN mv bin/${GOOS}-${GOARCH} binaries && cp $(which curl) binaries
RUN mkdir tmp_dir

FROM scratch
WORKDIR /
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=build /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=build /src/tmp_dir /tmp
COPY --from=build /src/binaries bin
COPY configs configs
COPY web web
# HEALTHCHECK --interval=30s --timeout=4s CMD curl -f http://localhost:8080/live || exit 1 @TODO:
EXPOSE 8080
ENTRYPOINT ["server"]
