FROM alpine:3.5

WORKDIR /
COPY ./platformer_service /

EXPOSE 80 443

VOLUME ["/certs"]

ENTRYPOINT ["./platformer_service"]

