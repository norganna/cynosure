#!/bin/bash

protoc \
  -I. \
  -I"${GOPATH}/src" \
  -I"${GOPATH}/src/github.com/grpc-ecosystem/grpc-gateway" \
  -I"${GOPATH}/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis" \
  --go_out=plugins=grpc:. \
  --grpc-gateway_out=logtostderr=true:. \
  --swagger_out=logtostderr=true:. \
  cyno.proto

go generate .
