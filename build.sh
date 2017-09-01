#!/bin/bash

cat > Dockerfile.builder <<EOF
FROM golang

MAINTAINER Denis Zheleztsov <difrex.punk@gmail.com>

RUN go get github.com/Difrex/rbmd/rbmd
RUN cd /go/src/github.com/Difrex/rbmd && go get -t -v ./... || true

WORKDIR /go/src/github.com/Difrex/rbmd

ENTRYPOINT go build -ldflags "-linkmode external -extldflags -static" -o rbmd-linux-amd64 && mv rbmd-linux-amd64 /out
EOF

# Build builder
docker build --no-cache -t rbmd_builder -f Dockerfile.builder .
# Build bin
docker run -v $(pwd)/out:/out rbmd_builder

