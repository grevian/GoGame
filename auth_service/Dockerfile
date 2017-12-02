FROM alpine:3.5

WORKDIR /
COPY /workspace/gopath/bin/auth_service /

EXPOSE 80 443

VOLUME ["/certs"]

ENTRYPOINT ["./auth_service"]
