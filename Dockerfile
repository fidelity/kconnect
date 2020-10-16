FROM gcr.io/distroless/static:latest

COPY kconnect /

USER kconnect

ENTRYPOINT ["/kconnect"]
