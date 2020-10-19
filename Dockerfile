FROM gcr.io/distroless/static:latest

USER root
COPY --chown=nonroot:nonroot kconnect /

USER nonroot
ENTRYPOINT ["/kconnect"]
