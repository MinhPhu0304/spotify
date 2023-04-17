#!/bin/bash

echo "How does this work, I don't know (╯ಠ‿ಠ )╯︵┻━┻ "
echo "Build for linux architecture"

GOOS=linux GOARCH=amd64 go build -o main main.go

echo "Zip stuff toghether"
zip main.zip main

chmod 777 main.zip

echo ****** Updating "function code " *******

aws lambda update-function-code \
    --function-name  spotify \
    --zip-file fileb://main.zip --publish | jq '
  if .Environment.Variables.SPOTIFY_SECRET? then .Environment.Variables.SPOTIFY_SECRET = "REDACTED" else . end |
  if .Environment.Variables.CALLBACK_URI? then .Environment.Variables.CALLBACK_URI = "REDACTED" else . end |
  if .Environment.Variables.SENTRY_DSN? then .Environment.Variables.SENTRY_DSN = "REDACTED" else . end |
  if .Environment.Variables.LASTFM_API_KEY? then .Environment.Variables.LASTFM_API_KEY = "REDACTED" else . end |
  if .Environment.Variables.SPOTIFY_ID? then .Environment.Variables.SPOTIFY_ID = "REDACTED" else . end' 

echo \n ******** All done. Check prod *************