#!/bin/bash

rm -rf ./packages/file-service/protogen/*

protoc --go_out=./packages/file-service/protogen \
    --go_opt=paths=source_relative \
    --go-grpc_out=./packages/file-service/protogen \
    --go-grpc_opt=paths=source_relative -I ./proto file-service.proto
