#!/bin/bash


echo "Build for linux architecture"

GOOS=linux GOARCH=amd64 go build -o main main.go

echo "Zip stuff toghether"
zip main.zip main

chmod 777 main.zip

echo "updating function code"

aws lambda update-function-code \
    --function-name  spotify \
    --zip-file fileb://main.zip --publish

echo "all done. Check prod"