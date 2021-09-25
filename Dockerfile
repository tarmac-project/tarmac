FROM golang:1.16.8

ADD . /go/src/github.com/madflojo/tarmac
WORKDIR /go/src/github.com/madflojo/tarmac/cmd/tarmac
RUN go install -v .
WORKDIR /go/src/github.com/madflojo/tarmac/

ENTRYPOINT ["./docker-entrypoint.sh"]
