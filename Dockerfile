FROM golang:alpine

ARG version
ARG go_get_http_proxy

# Download ca-certificates & git from aliyun mirrors
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories
RUN apk --update add --no-cache ca-certificates git

COPY . /build/iap-gateway/
WORKDIR /build/iap-gateway/
RUN http_proxy=$go_get_http_proxy https_proxy=$go_get_http_proxy go get -v ./... || exit 0
RUN GOOS=linux CGO_ENABLED=0 go build -ldflags "-X main.Version=$version"

FROM scratch

COPY --from=0 /usr/share/ca-certificates /usr/share/ca-certificates
COPY --from=0 /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=0 /etc/passwd /etc/passwd
COPY --from=0 /build/iap-gateway/iap-gateway /bin/

WORKDIR /

EXPOSE 8091
USER nobody

CMD ["/bin/iap-gateway", "-config", "/etc/iap-gateway/iap-gateway.yml"]
