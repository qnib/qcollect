#!/bin/bash
set -x

cp main.go main.go.bkp
GIT_TAG=$(git describe --abbrev=0 --tags)
if [ -f /etc/os-release ];then
    . /etc/os-release
    if [ "X${ID}" != "Xalpine" ];then
      ID=Linux
      sed -i'' -e "s/version =.*/version = \"${GIT_TAG}\"/" main.go
    else
      sed -i '' -e 's/version =.*/version = "0.6.4.1"/' main.go
    fi
else
    ID=$(uname -s)
    sed -i '' -e "s/version =.*/version = \"${GIT_TAG}\"/" main.go
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
go build -o ./bin/qcollect_${GIT_TAG}_${ID}
mv main.go.bkp main.go
rm -f ./bin/qcollect_latest_${ID}
cp ./bin/qcollect_${GIT_TAG}_${ID} ./bin/qcollect_latest_${ID}
