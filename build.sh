#!/bin/bash

#Linux
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ob_linux

#Mac
CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o ob_mac_intel
CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -o ob_mac_arm

#Windows
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o ob_win.exe