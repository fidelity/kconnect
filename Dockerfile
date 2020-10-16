FROM scratch

COPY kconnect /
USER kconnect
ENTRYPOINT ["/kconnect"]
