FROM golang:1.26-alpine AS build
WORKDIR /go/src/github.com/utilitywarehouse/gcp-disk-snapshotter
COPY . /go/src/github.com/utilitywarehouse/gcp-disk-snapshotter
ENV CGO_ENABLED=0
# GOTOOLCHAIN pins the exact toolchain declared by go.mod's `go` line, so the
# build isn't at the mercy of whatever patch version the base image ships.
RUN apk --no-cache add git &&\
  GOTOOLCHAIN=go$(awk '/^go /{print $2; exit}' go.mod) && \
  go mod download &&\
  go test ./... &&\
  go build -o /gcp-disk-snapshotter .

FROM alpine:3.24
RUN apk --no-cache add ca-certificates
COPY --from=build /gcp-disk-snapshotter /gcp-disk-snapshotter
ENTRYPOINT [ "/gcp-disk-snapshotter" ]
