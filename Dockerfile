# syntax=docker/dockerfile:1.0

ARG GOPATH=/go
ARG PKG=github.com/benshields/messagebox
ARG GODIR=$GOPATH/src/$PKG

# Build
FROM golang:1.16-buster AS build
ARG PKG
ARG GODIR
WORKDIR $GODIR
COPY ./ $GODIR
RUN go build -v -o /go/bin/messagebox ${PKG}/cmd/api

# Deploy
FROM gcr.io/distroless/base-debian10
ARG GODIR
COPY --from=build /go/bin/messagebox /go/bin/messagebox
COPY --from=build $GODIR/config/default.yaml /go/bin/config/
WORKDIR /go/bin
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/go/bin/messagebox"]
