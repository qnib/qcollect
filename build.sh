#!/bin/bash
set -x

if [ -f /etc/os-release ];then
    . /etc/os-release
    if [ "X${ID}" != "Xalpine" ];then
      ID=Linux
    fi
else
    ID=$(uname -s)
fi
if [ ! -d ${GOPATH}/src/github.com/davecheney/profile ];then
    git clone https://github.com/davecheney/profile.git ${GOPATH}/src/github.com/davecheney/profile
fi
go get -d
pushd ${GOPATH}/src/github.com/davecheney/profile
git checkout v0.1.0-rc.1
popd
go build -o qcollect_$(git describe --abbrev=0 --tags)_${ID}
