FROM golang:1.18 as debug
# RUN go get \
#         github.com/swaggo/swag/cmd/swag \
#         github.com/matryer/moq \
#         2>&1
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

FROM scratch as server
WORKDIR /app
COPY --from=build /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=build /src/bin bin
COPY configs configs
COPY web web
VOLUME [ "/app/storage" ]
ENTRYPOINT ["bin/server"]
