#!/bin/sh

go mod tidy
CGO_ENABLED=0 go build -ldflags '-s -w -extldflags "-static"' -trimpath .
