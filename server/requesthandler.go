package main

import (
	"net"
	"strings"
)

//handles requests for the server
func HandleNewRequest() (net.Conn, error) {
	var tcpl net.Listener
	var conn net.Conn
	var msg Netmessage
	var err error
	tcpl, err = net.Listen(TCP, IP_SERVER)
	Errhandle_Log(err, ERRMSG_TCPLISTENER)
	if err != nil {
		return conn, err
	}
	conn, err = tcpl.Accept()
	Errhandle_Log(err, ERRMSG_NETWORK_CONNECTION)
	if err != nil {
		return conn, err
	}
	err = ReceiveStruct(&msg, conn)
	Errhandle_Log(err, ERRMSG_NETWORK_RECEIVE_STRUCT)
	if err != nil {
		return conn, err
	}
	switch msg.Message {
	case NETREQ_NEWDEVICE:
		RequestDevice(conn)
	}
	return conn, nil
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
