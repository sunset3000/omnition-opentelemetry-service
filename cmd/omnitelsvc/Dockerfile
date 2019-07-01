FROM alpine:latest as certs
RUN apk --update add ca-certificates

FROM scratch
COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY omnitelsvc /usr/bin/
ENTRYPOINT ["/usr/bin/omnitelsvc"]
EXPOSE 55678
