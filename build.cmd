@echo off
SET CGO_ENABLED=0
SET GOOS=linux
SET GOARCH=amd64
go build -o bin/vnstat_web_linux_amd64
SET CGO_ENABLED=0
SET GOOS=linux
SET GOARCH=arm64
go build -o bin/vnstat_web_linux_arm64
SET CGO_ENABLED=0
SET GOOS=linux
SET GOARCH=arm
go build -o bin/vnstat_web_linux_arm