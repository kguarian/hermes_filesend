package main

import (
	"C"
	"net"
)
import (
	"fmt"
)

func main() {}

//export DeviceConn
func DeviceConn(userid *C.char, devicename *C.char) {
	var devinf deviceinfo
	var received_device device
	var connServer net.Conn
	var err error
	var msg = Netmessage{NETREQ_NEWDEVICE}
	var uid string = C.GoString(userid)
	var dname string = C.GoString(devicename)

	devinf = deviceinfo{Userid: uid, Devicename: dname}
	connServer, err = net.Dial(TCP, IP_SERVER)
	Errhandle_Log(err, ERRMSG_NETWORK_DIAL)
	if err != nil {
		return
	}
	defer connServer.Close()
	Errhandle_Log(err, ERRMSG_NETWORK_SEND_STRUCT)
	if err != nil {
		return
	}
	err = SendStruct(&msg, connServer)
	Errhandle_Log(err, ERRMSG_NETWORK_SEND_STRUCT)
	if err != nil {
		return
	}
	err = ReceiveStruct(&msg, connServer)
	Errhandle_Log(err, ERRMSG_NETWORK_READ)
	if err != nil {
		return
	}
	err = SendStruct(&devinf, connServer)
	Errhandle_Log(err, ERRMSG_NETWORK_SEND_STRUCT)
	if err != nil {
		return
	}
	err = ReceiveStruct(&received_device, connServer)
	Errhandle_Log(err, ERRMSG_NETWORK_WRITE)
	if err != nil {
		return
	}
	fmt.Printf("received device: %v\n", received_device)
}
