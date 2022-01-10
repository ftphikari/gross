#!/bin/sh

go mod tidy
CGO_ENABLED=0 go build -o gross.out -ldflags '-s -w -extldflags "-static"' -trimpath .
