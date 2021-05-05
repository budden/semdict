FROM ubuntu:bionic as configure
    RUN apt-get update
    RUN apt-get install -y curl gnupg2 lsb-release
    RUN curl -L https://packagecloud.io/golang-migrate/migrate/gpgkey | apt-key add -
    RUN echo "deb https://packagecloud.io/golang-migrate/migrate/ubuntu/ $(lsb_release -sc) main" > /etc/apt/sources.list.d/migrate.list
    RUN apt-get update
    RUN apt-get install -y migrate
    RUN migrate --version

    RUN apt install -y postgresql-client
    RUN psql --version

    WORKDIR /

FROM golang:1.14 as semdict-server-builder
    WORKDIR /go/src/github.com/budden/semdict
    COPY go.mod .
    COPY go.sum .
    RUN go mod download
    COPY . .
    RUN go build -v -o /semdict-server ./main.go

FROM alpine:3.11 as semdict-server
    RUN apk update && apk add --no-cache git ca-certificates tzdata make dbus && update-ca-certificates
    RUN dbus-uuidgen > /etc/machine-id
    RUN mkdir /lib64 && ln -s /lib/libc.musl-x86_64.so.1 /lib64/ld-linux-x86-64.so.2
    WORKDIR /
    COPY --from=semdict-server-builder /semdict-server .
    COPY --from=semdict-server-builder /go/src/github.com/budden/semdict/templates ./templates
    COPY --from=semdict-server-builder /go/src/github.com/budden/semdict/static ./static
    ENTRYPOINT /semdict-server






