package main

import (
	// #include <sys/types.h>
	// #include <sys/socket.h>
	"C"
	"net"
)
import (
	"flag"
	"log"
)

// //export DeviceConn
// func DeviceConn(userid string, devicename string) {
// 	var sentdev DeviceInfo
// 	var connmain net.Conn
// 	var msg []byte
// 	var err error
// 	sentdev = DeviceInfo{Userid: userid, Devicename: devicename}
// 	connmain, err = net.Dial(TCP, IP_main)
// 	Errhandle_Log(err, ERRMSG_NETWORK_DIAL)
// 	if err != nil {
// 		return
// 	}
// 	defer connmain.Close()
// 	msg, err = json.Marshal(sentdev)
// 	Errhandle_Log(err, ERRMSG_NETWORK_DIAL)
// 	if err != nil {
// 		return
// 	}
// 	connmain.Write(msg)
// }

func main() {
	var errorchannel chan error
	var err error
	logflag = flag.Bool("log", true, USTRING_SHOWLOG)
	awayflag = flag.Bool("away", false, USTRING_AWAY)
	flag.Parse()

	_, errorchannel = InitiateEverything()
	log.Printf("channel length = %d", len(errorchannel))
	for len(errorchannel) != 0 {
		err = <-errorchannel
		Errhandle_Exit(err, err.Error())
	}
	tcpl, err := net.Listen(TCP, IP_main) //change
	Errhandle_Exit(err, ERRMSG_NETWORK_CONNECTION)
	for {
		//fmt.Printf("%sRESTARTING HERE\n%s", ANSIRED, ANSIRESET)
		conn, err := tcpl.Accept()
		Errhandle_Log(err, ERRMSG_NETWORK_CONNECTION)
		go Ambassador(conn)
		//fmt.Printf("%sFINISHED HERE\n%s", ANSIRED, ANSIRESET)
	}
}

//This function should always work
func tesfunc() {

	InitDeviceTables()

	// dev0, err := NewDevice("kguarian", "dev0", net.ParseIP(IP_DEV0))
	// Errhandle_Exit(err, ERRMSG_CREATE_DEVICE)

	// dev1, err := NewDevice("kguarian", "dev1", net.ParseIP(IP_DEV1))
	// Errhandle_Exit(err, ERRMSG_CREATE_DEVICE)

	// content := NewContent("testfile.txt")
	// content2 := NewContent("test2.txt")

	// cinf, err := dev0.NewContentinfo(&dev1, &content)

	// Errhandle_Exit(err, ERRMSG_SELFSEND)

	// cinf2, err := dev0.NewContentinfo(&dev1, &content2)

	// Errhandle_Exit(err, ERRMSG_SELFSEND)

	// dev1.Online = true

	// c, err := HandleNewRequest()
	// Errhandle_Exit(err, ERRMSG_NETWORK_CONNECTION)
	// defer c.Close()
	// err = dev0.RequestConsent(dev1, cinf)
	// Errhandle_Exit(err, ERRMSG_DEVICEOFFLINE)
	// err = dev0.RequestConsent(dev1, cinf2)
	// Errhandle_Exit(err, ERRMSG_DEVICEOFFLINE)

	// err = dev1.EvalConsent(&dev0)
	// Errhandle_Exit(err, ERRMSG_CHANNEL_OPERATION)
	// fmt.Println(dev1.Consentlist)

	// SetConsoleColor(GREEN)
	// fmt.Println("Send, and everyone's happy!")
	// SetConsoleColor(RESET)

	//----------

	for {
		tcpl, err := net.Listen(TCP, IP_main)
		Errhandle_Log(err, ERRMSG_NETWORK_CONNECTION)
		conn, err := tcpl.Accept()
		Errhandle_Log(err, ERRMSG_NETWORK_CONNECTION)
		go Ambassador(conn)
	}

}
