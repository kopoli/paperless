#!/bin/sh

die() {
    echo "Error: $@" 1>&2
    exit 1
}

PAPERLESS=../paperless
IMG=ubuntu-16.04-paperless:latest
DATADIR=../docker-data

set -x

test -e "$PAPERLESS" || die "Build paperless before trying to run this"

mkdir -p $DATADIR

docker build . -t $IMG || die "Could not build the docker image"

docker run -v $PWD/$PAPERLESS:/paperless -v $PWD/$DATADIR:/data \
       --rm \
       $IMG
