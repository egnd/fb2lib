FROM alpine:latest as builder
WORKDIR /src
RUN apk add -q tzdata && mkdir tmp

FROM scratch
ARG TARGETOS TARGETARCH
ENV GOOS=$TARGETOS
ENV GOARCH=$TARGETARCH
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /src/tmp tmp
COPY bin/${GOOS}-${GOARCH} bin
COPY configs configs
COPY web web
VOLUME ["/var/index"]
ENTRYPOINT ["server"]
