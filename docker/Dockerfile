### API Build
FROM golang:alpine AS build
ENV CGO_ENABLED 0
WORKDIR /go/src/sambaapi
ADD ./API /go/src/sambaapi/
RUN go build .



FROM alpine

## API
COPY --from=build /go/src/sambaapi/SambaAPI /usr/bin

### samba
RUN apk --no-cache --no-progress upgrade && \
    apk --no-cache --no-progress add bash samba shadow tini tzdata samba-dc krb5 ca-certificates samba-winbind-clients

EXPOSE 137/udp 138/udp 139 445

COPY ./docker/resolv.conf /etc

VOLUME ["/etc", "/var/cache/samba", "/var/lib/samba", "/var/log/samba", "/run/samba"]

## health check
HEALTHCHECK --interval=60s --timeout=15s \
            CMD smbclient -L \\localhost -U % -m SMB3 && wget -O /dev/null http://localhost:8081/ping

## entrypoint
COPY ./docker/entrypoint.sh /usr/bin
ENTRYPOINT [ "bash", "/usr/bin/entrypoint.sh" ]