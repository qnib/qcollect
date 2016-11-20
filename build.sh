#!/bin/bash

cp main.go main.go.bkp
GIT_ORG_TAG=$(git describe --abbrev=0 --tags)
git describe --exact-match --abbrev=0 > /dev/null
if [ $? -ne 0 ];then
    GIT_TAG="${GIT_ORG_TAG}-dirty"
    BC_CMD=$(which bc)
    if [ $? -ne 0 ];then
        echo "!! Need bc command to calculate the number of commits since the latest tag -> proceed without..."
    else
        ## commit since tags
        CNT_ALL=$(git log --oneline |wc -l)
        CNT_COMMITS=$(echo "${CNT_ALL}-$(git log --oneline ${GIT_ORG_TAG} |wc -l)" |bc)
        if [ ${CNT_COMMITS} -ne 0 ];then
            GIT_TAG="${GIT_TAG}-${CNT_COMMITS}"
        else
            GIT_TAG="${GIT_ORG_TAG}"
        fi
    fi
fi

if [ -f /etc/os-release ];then
    . /etc/os-release
    if [ "X${ID}" != "Xalpine" ];then
      ID=Linux
      sed -i'' -e "s/version =.*/version = \"${GIT_TAG}\"/" main.go
    else
      sed -i'' -e "s/version =.*/version = \"${GIT_TAG}\"/" main.go
    fi
else
    ID=$(uname -s)
    sed -i '' -e "s/version =.*/version = \"${GIT_TAG}\"/" main.go
fi

if [ ! -d ${GOPATH}/src/github.com/davecheney/profile ];then
    git clone https://github.com/davecheney/profile.git ${GOPATH}/src/github.com/davecheney/profile
fi

govendor sync

rm -f ./bin/qcollect_${GIT_TAG}_${ID}
go build -o ./bin/qcollect_${GIT_TAG}_${ID}
mv main.go.bkp main.go
rm -f ./bin/qcollect_latest_${ID}
cp ./bin/qcollect_${GIT_TAG}_${ID} ./bin/qcollect_latest_${ID}
