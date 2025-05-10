#!/bin/bash

# gen proto for file-service
function gen-file-service-proto() {
    rm -rf ./packages/file-service/protogen/*

    protoc --go_out=./packages/file-service/protogen \
        --go_opt=paths=source_relative \
        --go-grpc_out=./packages/file-service/protogen \
        --go-grpc_opt=paths=source_relative -I ./proto file-service.proto
}

# gen proto for gateway
function gen-gateway-proto() {
    rm -rf ./packages/gateway/protogen/*

    protoc --go_out=./packages/gateway/protogen \
        --go_opt=paths=source_relative \
        --go-grpc_out=./packages/gateway/protogen \
        --go-grpc_opt=paths=source_relative -I ./proto file-service.proto

    protoc --go_out=./packages/gateway/protogen \
        --go_opt=paths=source_relative \
        --go-grpc_out=./packages/gateway/protogen \
        --go-grpc_opt=paths=source_relative -I ./proto auth-service.proto
}

function gen-auth-service-proto() {
    rm -rf ./packages/auth-service/protogen/*

    protoc --go_out=./packages/auth-service/protogen \
        --go_opt=paths=source_relative \
        --go-grpc_out=./packages/auth-service/protogen \
        --go-grpc_opt=paths=source_relative -I ./proto auth-service.proto
}

case $1 in
"file-service")
    gen-file-service-proto
    ;;
"gateway")
    gen-gateway-proto
    ;;
"auth-service")
    gen-auth-service-proto
    ;;
*) # going to re-build proto for all packages
    gen-file-service-proto
    gen-gateway-proto
    gen-auth-service-proto
    ;;
esac
