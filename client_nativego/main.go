package main

import (
	"C"
)

var functionmap map[string]func([]string) error = make(map[string]func([]string) error)

func main() {
	functionmap["devconn"] = DeviceConn
	print(DeviceConn([]string{"cppclient", "kguarian"}))
}

//export DeviceConn
func DeviceConn(args []string) error {
	//ARGS
	var uid = args[0]
	var dname = args[1]

	var devinf deviceinfo = deviceinfo{uid, dname}
	var received_device device
	var msg = Netmessage{NETREQ_NEWDEVICE}

	// //If you're away from home, uncomment this block.
	// err = selectServerIP()
	// Errhandle_Exit(err, ERRMSG_INPUT_RETRIEVAL)

	// fmt.Printf("IP_SERVER: %s\n", IP_SERVER)

	err := PassRequest(&msg, &devinf, &received_device)
	return err
}

// func LinkRequest(args []string) {
// 	var tgtuser string = args[0]

// }
