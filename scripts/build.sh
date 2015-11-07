#!/bin/sh

cd `dirname $0`

GOOS=linux GOARCH=amd64 go build -o ../src/NewVideoCollector ../src/NewVideoCollector.go
GOOS=linux GOARCH=amd64 go build -o ../src/VideoAnalyzer ../src/VideoAnalyzer.go
