FROM alpine:3.5

WORKDIR /
COPY gopath/bin/server /

EXPOSE 8077 8078

VOLUME ["/certs"]

ENTRYPOINT ["./server"]

