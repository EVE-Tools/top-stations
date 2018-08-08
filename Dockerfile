#
# Build project in separate container
#

FROM golang:alpine3.8 AS build

RUN apk update && \
    apk upgrade && \
    apk add git
COPY . /go/src/github.com/EVE-Tools/top-stations

WORKDIR /go/src/github.com/EVE-Tools/top-stations
RUN go get -v ./...
RUN go build
RUN cp /go/src/github.com/EVE-Tools/top-stations/top-stations /top-stations

#
# Copy release to fresh container and set command
#

FROM alpine:3.8

# Update base system, load ca certs, otherwise no TLS for us
RUN apk update && \
    apk upgrade && \
    apk add ca-certificates && \
    rm -rf /var/cache/apk/*

# Do not run as root
RUN addgroup -g 1000 -S element43 && \
    adduser -u 1000 -S element43 -G element43 && \
    mkdir /data && \
    chown -R element43:element43 /data
USER element43:element43

# Copy build
COPY --from=build /top-stations /top-stations

ENV PORT 43000
EXPOSE 43000

CMD ["/top-stations"]