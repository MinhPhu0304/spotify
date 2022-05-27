#!/bin/bash

GOOS=linux GOARCH=amd64 go build -o main main.go
zip main.zip main
chmod 777 main.zip