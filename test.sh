#!/bin/bash
set -ex

govendor fetch +missing
echo "> govendor remove +unused"
govendor remove +unused
echo "> govendor sync"
govendor sync
if [ ! -d resources/coverity ];then
    mkdir -p resources/coverity
fi
go test -coverprofile=coverage.out >/dev/null
COVER_OUTS="coverage.out"
for x in $(find . -maxdepth 1 -type d |egrep -v "(\.$|\.git|vendor|bin|resources|deploy)");do
    go test -coverprofile=resources/coverity/${x}.out ${x} >/dev/null
    COVER_OUTS="${COVER_OUTS} resources/coverity/${x}.out"
done
coveraggregator -o resources/coverity/coverage-all.out ${COVER_OUTS} >/dev/null
go tool cover -func=resources/coverity/coverage-all.out |tee ./coverage-all.out
go tool cover -html=resources/coverity/coverage-all.out -o resources/coverity/all.html
