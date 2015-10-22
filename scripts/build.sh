#!/bin/sh

cd `dirname $0`

GOOS=linux GOARCH=amd64 go build ../src/NewVideoCollector.go
GOOS=linux GOARCH=amd64 go build ../src/VideoAnalyzer.go
