FROM golang:1.25 AS builder

WORKDIR /src

COPY ./go.mod ./
COPY ./go.sum ./
RUN go mod download

COPY ./cmd ./cmd
COPY ./internal ./internal
COPY ./pkg ./pkg

RUN CGO_ENABLED=0 go build -o /server ./cmd/server

FROM scratch
# CA証明書をコピー
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /server /server
CMD ["/server"]
