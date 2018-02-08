#!/bin/bash
mkdir -p ./lib/topStations ./lib/esiMarkets

go get -u github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway
go get -u github.com/golang/protobuf/protoc-gen-go

protoc -I../element43/services/topStations \
-I$GOPATH/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
--go_out=plugins=grpc:./lib/topStations \
../element43/services/topStations/topStations.proto

protoc -I../element43/services/esiMarkets \
-I$GOPATH/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
--go_out=plugins=grpc:./lib/esiMarkets \
../element43/services/esiMarkets/esiMarkets.proto