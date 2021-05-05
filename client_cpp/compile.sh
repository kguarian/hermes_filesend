#!/usr/bin/bash
bash build/build.sh
cd cpp
gcc -pthread *.c *.a *.h
cd ../
./cpp/a.out