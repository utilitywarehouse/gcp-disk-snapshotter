FROM alpine:3.7

ENV GOPATH=/go

WORKDIR /go/src/github.com/utilitywarehouse/gcp-disk-snapshotter
COPY . /go/src/github.com/utilitywarehouse/gcp-disk-snapshotter

RUN \
 apk --no-cache add ca-certificates git go musl-dev && \
 go get -t ./... && \
 go test ./... && \
 CGO_ENABLED=0 go build -ldflags '-s -extldflags "-static"' -o /gcp-disk-snapshotter . && \
 apk del go musl-dev && rm -r /go

ENTRYPOINT [ "/gcp-disk-snapshotter" ]
