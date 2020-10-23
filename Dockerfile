FROM scratch

COPY  kconnect /

ENTRYPOINT ["/kconnect"]
