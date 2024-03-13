#!/bin/sh
cd /go/src/in_twsnmp
go build -buildmode=c-shared -o in_twsnmp.linux.amd64.so .
CC=aarch64-linux-gnu-gcc GOOS=linux GOARCH=arm64  CGO_ENABLED=1 go build -buildmode=c-shared -o in_twsnmp.linux.arm64.so 
CC=arm-linux-gnueabihf-gcc GOOS=linux GOARCH=arm  CGO_ENABLED=1 go build -buildmode=c-shared -o in_twsnmp.linux.arm.so 
GOOS=windows GOARCH=amd64 CGO_ENABLED=1 CXX=x86_64-w64-mingw32-g++ CC=x86_64-w64-mingw32-gcc go build -buildmode=c-shared -o in_twsnmp.windows.dll .
