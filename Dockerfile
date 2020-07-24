FROM golang:1.14-alpine AS build
WORKDIR /go/src/github.com/utilitywarehouse/gcp-disk-snapshotter
COPY . /go/src/github.com/utilitywarehouse/gcp-disk-snapshotter
ENV CGO_ENABLED 0
RUN apk --no-cache add git &&\
  go get -t ./... &&\
  go test ./... &&\
  go build -o /gcp-disk-snapshotter .

FROM alpine:3.12
RUN apk --no-cache add ca-certificates
COPY --from=build /gcp-disk-snapshotter /gcp-disk-snapshotter
ENTRYPOINT [ "/gcp-disk-snapshotter" ]
