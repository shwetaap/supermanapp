FROM golang:1.12-stretch as builder

WORKDIR /build/myapp

# Fetch dependencies
COPY go.mod go.sum ./
RUN go mod download

# Build
COPY . ./
RUN CGO_ENABLED=1 go build -mod=vendor

# Create final image
FROM alpine
RUN mkdir /lib64 && ln -s /lib/libc.musl-x86_64.so.1 /lib64/ld-linux-x86-64.so.2

WORKDIR /root

COPY --from=builder build/myapp/supermanapp .
COPY GeoLite2-City.mmdb .
EXPOSE 8000
CMD ["./supermanapp"]