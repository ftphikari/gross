#!/bin/sh

go mod tidy
CGO_ENABLED=0 go build -o gross.elf -ldflags '-s -w -extldflags "-static"' -trimpath .
