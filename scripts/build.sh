#!/bin/sh

cd `dirname $0`

GOOS=linux GOARCH=amd64 go build -o ../src/ ../src/NewVideoCollector.go
GOOS=linux GOARCH=amd64 go build -o ../src/ ../src/VideoAnalyzer.go
