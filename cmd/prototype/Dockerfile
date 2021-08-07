# Source: https://levelup.gitconnected.com/complete-guide-to-create-docker-container-for-your-golang-application-80f3fb59a15e
FROM golang:1.16-alpine AS builder

RUN apk add -U --no-cache ca-certificates

# Set necessary environmet variables needed for our image
ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

ADD . /dockerdev
WORKDIR /dockerdev

RUN go build -o /prototype ./cmd/prototype

# Build a small image
FROM scratch

EXPOSE 8000

WORKDIR /
COPY --from=builder /prototype /
COPY --from=alpine /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

ENTRYPOINT ["/prototype"]