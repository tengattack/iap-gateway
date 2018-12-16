FROM alpine:latest

# Download ca-certificates from aliyun mirrors
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories
RUN apk --update add --no-cache ca-certificates

FROM scratch

COPY --from=0 /usr/share/ca-certificates /usr/share/ca-certificates
COPY --from=0 /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=0 /etc/passwd /etc/passwd
COPY iap-gateway /bin/

WORKDIR /

EXPOSE 8091
USER nobody

CMD ["/bin/iap-gateway", "-config", "/etc/iap-gateway/iap-gateway.yml"]
