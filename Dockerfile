FROM golang:latest

ADD . /go/src/github.com/tarmac-project/tarmac
WORKDIR /go/src/github.com/tarmac-project/tarmac/
RUN go mod tidy
WORKDIR /go/src/github.com/tarmac-project/tarmac/cmd/tarmac
RUN go install -v .
WORKDIR /go/src/github.com/tarmac-project/tarmac/

FROM ubuntu:latest
# Install latest CA certificates
RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/* && update-ca-certificates
RUN install -d -m 0755 -o 1000 -g 500 /app/tarmac
# Create Data directory for local data storage, override with volume mounts to retain data
RUN install -d -m 0755 -o 1000 -g 500 /data/tarmac
COPY --chown=1000:500 --from=0 /go/bin/tarmac /app/tarmac/
COPY --chown=1000:500 --from=0 /go/src/github.com/tarmac-project/tarmac/docker-entrypoint.sh /app/tarmac/
RUN chmod 755 /app/tarmac/tarmac /app/tarmac/docker-entrypoint.sh
USER 1000

ENTRYPOINT ["/app/tarmac/docker-entrypoint.sh"]
