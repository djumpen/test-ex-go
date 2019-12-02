#!/bin/bash
# dep ensure
rm -r ./bin
mkdir ./bin
go build -o ./bin/app cmd/*.go
bin/app
