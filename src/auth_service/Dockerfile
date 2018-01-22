FROM alpine:3.5

WORKDIR /
COPY ./auth_service /

EXPOSE 80 443

VOLUME ["/certs"]

ENTRYPOINT ["./auth_service"]
