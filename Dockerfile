FROM alpine:3.5

WORKDIR /
COPY certs/ /
COPY gopath/bin/server /

ENTRYPOINT ["./server"]

