#!/usr/bin/bash
WD=$(pwd)
cd /home/guarian/HOME/coding/go/src/rnd/hermes_filesend/server;
sed -i 's/main/server/g' *.go;
go build server;
cd ../;
cd hellabackend;
go build hellabackend;
cd ../;
sudo rm -r /usr/local/go/src/hermes/*;
sudo cp -r server/ /usr/local/go/src/hermes/server;
#Cannot have main.go file
sudo rm /usr/local/go/src/hermes/server/main.go
sudo cp -r hellabackend /usr/local/go/src/hermes/;


#cleanup
cd server;
sed -i 's/server/main/g' *.go;
cd ../

cd $WD;