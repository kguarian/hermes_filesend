#!/usr/bin/bash

go build -buildmode=c-archive *.go
mv *.a *.h cpp/
cd cpp
gcc -pthread *.c *.a *.h
cd ..