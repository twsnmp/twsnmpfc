#!/bin/sh
apt update
apt install -y nodejs npm python libx11-dev libxkbfile-dev libxkbfile-dev  xserver-xorg-dev libxi-dev libxext-dev
go install github.com/rakyll/statik
