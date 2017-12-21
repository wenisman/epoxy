## builder for the docker deployed binaries
FROM golang:latest as builder
WORKDIR /go/src/github.com/wenisman/epoxy
RUN go get -d -v golang.org/x/net/html
COPY . .
RUN go get -v
RUN go test -v ./lib/*
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags "-s" -a -installsuffix cgo .

## build the docker image using the built binaries
FROM alpine:latest
ENV PROXY_PORT 9001
ENV PROXY_PROXIES ""
ENV ENVIRONMENT ""

RUN apk --no-cache add ca-certificates

RUN mkdir -p /opt/epoxy
WORKDIR /opt/epoxy

COPY --from=builder /go/src/github.com/wenisman/epoxy/epoxy .
COPY --from=builder /go/src/github.com/wenisman/epoxy/config ./config

VOLUME [ "/opt/epoxy/config" ]

EXPOSE ${PROXY_PORT}

CMD ["/opt/epoxy/epoxy", "start"]