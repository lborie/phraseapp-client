#!/bin/bash
set -xe

BUILD_INFO=$(bash ./build_info.sh)
if [[ -z $BUILD_INFO ]]; then
	echo "BUILD_INFO was empty"
	exit 1
fi

d=$(mktemp -d /tmp/phrase-client-XXXXX)
trap "rm -Rf $d" EXIT

echo "building in $d"

go get -d ./...
export GOBIN=$d
go install -ldflags "-X main.BUILD_INFO $BUILD_INFO" ./...

install $d/phraseapp-client $HOME/bin/phraseapp-client
