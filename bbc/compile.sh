#!/usr/bin/bash
GOOS=js GOARCH=wasm go build -o go_devconn.wasm bbc

sudo cp go_devconn.wasm index.html /var/www/html/test/
