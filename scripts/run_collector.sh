#!/bin/sh

cd `dirname $0`
cd ..

./src/NewVideoCollector >> ./collector.log
