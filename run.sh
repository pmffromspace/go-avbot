#!/bin/sh

set -e
cd /go-avbot/
ls -l
pwd
ls -l ./data
go run init.go app.go $@
