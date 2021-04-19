package main

import (
	"net"
	"strings"
)

//handles requests for the server
func Ambassador(conn net.Conn) {
	var msg Netmessage
	var err error
	Errhandle_Log(err, ERRMSG_TCPLISTENER)
	if err != nil {
		return
	}
	err = ReceiveStruct(&msg, conn)
	Errhandle_Log(err, ERRMSG_NETWORK_RECEIVE_STRUCT)
	if err != nil {
		return
	}
	switch msg.Message {
	case NETREQ_NEWDEVICE:
		RequestDevice(conn)
	}

}

//Request to add/register device.
func RequestDevice(conn net.Conn) (device, error) {
	var retdevice device
	var ipstring string
	var ip net.IP
	var cdinf deviceinfo
	var msg Netmessage
	var err error

	msg = NewNetmessage(NETREQ_NEWDEVICE)
	err = SendStruct(&msg, conn)
	Errhandle_Log(err, ERRMSG_NETWORK_SEND_STRUCT)
	if err != nil {
		return retdevice, err
	}
	err = ReceiveStruct(&cdinf, conn)
	Errhandle_Log(err, ERRMSG_NETWORK_RECEIVE_STRUCT)
	if err != nil {
		return retdevice, err
	}
	ipstring = strings.Split(conn.RemoteAddr().String(), ":")[0]
	ip = net.ParseIP(ipstring)
	retdevice, err = NewDevice(cdinf.Userid, cdinf.Devicename, ip)
	Errhandle_Log(err, ERRMSG_CREATE_DEVICE)
	err = SendStruct(&retdevice, conn)
	Errhandle_Log(err, ERRMSG_NETWORK_SEND_STRUCT)
	//return or return, because we can't return a nil device
	return retdevice, err
}
