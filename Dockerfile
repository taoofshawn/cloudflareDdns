FROM golang:1.26.5-bookworm AS builder
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /cloudflareDdns .

FROM gcr.io/distroless/static:nonroot AS runner
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /cloudflareDdns /
CMD ["/cloudflareDdns"]
