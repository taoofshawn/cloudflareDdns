FROM golang:1.26.5-bookworm AS builder
RUN git clone https://github.com/taoofshawn/cloudflareDdns.git /cloudflareDdns && \
    cd /cloudflareDdns && \
    go build

FROM gcr.io/distroless/base:nonroot AS runner
COPY --from=builder /cloudflareDdns/cloudflareDdns /
CMD ["/cloudflareDdns"]
