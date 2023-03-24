FROM golang:latest

ADD . /go/src/github.com/madflojo/tarmac
WORKDIR /go/src/github.com/madflojo/tarmac/
RUN go mod tidy
WORKDIR /go/src/github.com/madflojo/tarmac/cmd/tarmac
RUN go install -v .
WORKDIR /go/src/github.com/madflojo/tarmac/

FROM ubuntu:latest
RUN install -d -m 0755 -o 1000 -g 500 /app/tarmac
COPY --chown=1000:500 --chmod=0755 --from=0 /go/bin/tarmac /app/tarmac/
COPY --chown=1000:500 --chmod=0755 --from=0 /go/src/github.com/madflojo/tarmac/docker-entrypoint.sh /app/tarmac/
USER 1000

ENTRYPOINT ["/app/tarmac/docker-entrypoint.sh"]
