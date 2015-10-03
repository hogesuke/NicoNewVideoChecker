#!/bin/sh

cd /home/hogesuke/nicotune/NicoNewVideoChecker/
go run ./src/NewVideoCollector.go >> ./collector.log
