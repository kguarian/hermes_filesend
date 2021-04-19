//Networking constants

package main

const (
	IP_SERVER        = "192.168.1.118:3128"
	NETCODE_ERR      = byte(0)
	NETCODE_SUC      = byte(1)
	PORT_LOWER_BOUND = 49152
	PORT_UPPER_BOUND = 65535
	TCP              = "tcp"
)
const (
	//Requests from device
	NETREQ_NEWDEVICE          = "nd"
	NETREQ_NEWCONTENTTRANSFER = "nct"
)

const (
	//Directive from server
	NETDIR_SENDCONTENTINFO = "s_ci"
)
