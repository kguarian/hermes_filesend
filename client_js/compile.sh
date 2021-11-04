#!/usr/bin/bash
GOOS=js GOARCH=wasm go build -o hermes_gosrc.wasm jsclient
sudo cp hermes_gosrc.wasm /var/www/html/hint/