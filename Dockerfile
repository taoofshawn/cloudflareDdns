FROM golang:buster as builder
RUN git clone https://github.com/taoofshawn/cloudflareDdns.git /cloudflareDdns && \
    cd /cloudflareDdns && \
    go build

FROM gcr.io/distroless/base as runner
COPY --from=builder /cloudflareDdns/cloudflareDdns /
CMD ["/cloudflareDdns"]