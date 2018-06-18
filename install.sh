#!/bin/bash

set -e

if test -z $GOPATH; then
    export GOPATH=$HOME/go
fi

mkdir -p $GOPATH/src/github.com/Azure/buffalo-azure
cd $GOPATH/src/github.com/Azure/buffalo-azure
git clone https://github.com/Azure/buffalo-azure.git . -q

export head=$(git rev-parse HEAD)
export relTag=$(git tag --list --contains "$head" v*)

if test -z $relTag; then
    export relTag="$head"
fi

dep ensure
go install --ldflags "-X github.com/Azure/buffalo-azure/cmd.version=$relTag"

echo "Installed Buffalo-Azure Version $relTag"
