FROM golang:1.25 AS builder

# Build argumentsを定義
ARG TARGETPLATFORM
ARG BUILDPLATFORM

WORKDIR /src

COPY ./go.mod ./
COPY ./go.sum ./
RUN go mod download

COPY ./cmd ./cmd
COPY ./internal ./internal
COPY ./pkg ./pkg

# プラットフォームに応じてGOOSとGOARCHを設定
RUN case ${TARGETPLATFORM} in \
        "linux/amd64")  export GOOS=linux GOARCH=amd64 ;; \
        "linux/arm64")  export GOOS=linux GOARCH=arm64 ;; \
        "linux/arm/v7") export GOOS=linux GOARCH=arm GOARM=7 ;; \
        *) echo "Unsupported platform: ${TARGETPLATFORM}" && exit 1 ;; \
    esac && \
    CGO_ENABLED=0 go build -o /server ./cmd/server

FROM scratch
# CA証明書をコピー
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /server /server
CMD ["/server"]
