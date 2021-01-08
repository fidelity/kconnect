FROM alpine:3.12 AS certs
RUN apk --update add ca-certificates

FROM scratch

COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY  kconnect /

ENTRYPOINT ["/kconnect"]
