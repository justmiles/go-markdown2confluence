#!/bin/sh
cd $( dirname "$0" )
echo "Cleaning dist"
rm -rf dist/*
echo "Building"
/go/src/app/bin/goreleaser --snapshot --skip-publish
if [ -r dist/checksums.txt ] ; then
    ls -l dist/*.tar.gz
else
    echo >&2 "Build failed"
    exit 1
fi