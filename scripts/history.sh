#!/usr/bin/env bash

exec >> history.log
date
noproxy go run cmd/tnkinv.go --token token.token
