#!/bin/bash
set -x

GIT_TAG=$(git describe --abbrev=0 --tags)
mkdir -p coverity
gom test -cover ./... |grep coverage |sed -e 's#github.com/qnib/##' |awk '{print $2" "$5}' > ./resources/coverity/cover_cur.out
./cover.plt > ./resources/coverity/cover_${GIT_TAG}.png
mv ./resources/coverity/cover_cur.out ./resources/coverity/cover_${GIT_TAG}.out
