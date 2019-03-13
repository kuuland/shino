#!/bin/sh

GOOS=linux
GOARCH=amd64
CGO_ENABLED=0
go build
tar -zcf shino-$GOOS-$GOARCH.tar.gz shino

GOOS=darwin
tar -zcf shino-$GOOS-$GOARCH.tar.gz shino

GOOS=windows
tar -zcf shino-$GOOS-$GOARCH.tar.gz shino

rm -f shino