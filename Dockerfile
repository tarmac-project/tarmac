FROM golang:latest

ADD . /go/src/github.com/madflojo/tarmac
WORKDIR /go/src/github.com/madflojo/tarmac/
RUN go mod tidy
WORKDIR /go/src/github.com/madflojo/tarmac/cmd/tarmac
RUN go install -v .
WORKDIR /go/src/github.com/madflojo/tarmac/

ENTRYPOINT ["./docker-entrypoint.sh"]
