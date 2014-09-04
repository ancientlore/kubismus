#!/bin/bash

binder -package static web/* tpl/* > static/files.go

pushd static > /dev/null
go install
popd > /dev/null
