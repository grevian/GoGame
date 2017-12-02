FROM alpine:3.5

WORKDIR /
COPY gopath/bin/server /

EXPOSE 80 443

VOLUME ["/certs"]

ENTRYPOINT ["./server"]
