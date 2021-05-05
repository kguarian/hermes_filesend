#!/usr/bin/bash
GOOS=js GOARCH=wasm go build -o go_devconn.wasm jsclient
sudo cp go_devconn.wasm /var/www/html/hint/
