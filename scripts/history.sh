#!/usr/bin/env bash

exec >> $(dirname $0)/../history.log
date
go run $(dirname $0)/../cmd/tnkinv.go $@
