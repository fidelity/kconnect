FROM alpine:3.19 AS certs
RUN apk --update add ca-certificates
RUN adduser -D kconnect

FROM scratch

COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY  kconnect /

COPY --from=certs /etc/passwd /etc/passwd
COPY --from=certs /home /home
USER kconnect
ENTRYPOINT ["/kconnect"]
