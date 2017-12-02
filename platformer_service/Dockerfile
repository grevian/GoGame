FROM alpine:3.5

WORKDIR /
COPY gopath/bin/platformer /

EXPOSE 80 443

VOLUME ["/certs"]

ENTRYPOINT ["./platformer"]

