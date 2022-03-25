ARG VARIANT="1.18-bullseye"
FROM mcr.microsoft.com/vscode/devcontainers/go:0-${VARIANT}

RUN DEBIAN_FRONTEND=noninteractive apt-get update && \
    DEBIAN_FRONTEND=noninteractive apt-get -y install --no-install-recommends tzdata openssh-client git make grep ca-certificates
# RUN go get \
#         github.com/swaggo/swag/cmd/swag \
#         github.com/matryer/moq \
#         2>&1
