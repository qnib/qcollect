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
go get github.com/pkg/errors github.com/stretchr/testify/assert

#mkdir -p coverity
#gom test -cover ./... |grep coverage |sed -e 's#github.com/qnib/##' |awk '{print $2" "$5}' > ./resources/coverity/cover_cur.out
#./cover.plt > ./resources/coverity/cover_$(git describe --abbrev=0 --tags).png
#mv ./resources/coverity/cover_cur.out ./resources/coverity/cover_$(git describe --abbrev=0 --tags).out
rm -f ./bin/qcollect_$(git describe --abbrev=0 --tags)_${ID}
cp main.go main.go.bkp
GIT_TAG=$(git describe --abbrev=0 --tags)
sed -i '' -e "s/version =.*/version = \"${GIT_TAG}\"/" main.go
go build -o ./bin/qcollect_${GIT_TAG}_${ID}
mv main.go.bkp main.go
rm -f ./bin/qcollect_latest_${ID}
cp ./bin/qcollect_${GIT_TAG}_${ID} ./bin/qcollect_latest_${ID}
