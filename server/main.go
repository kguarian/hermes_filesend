package main

import (
	// #include <sys/types.h>
	// #include <sys/socket.h>
	"C"
	"net"
)
import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"
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
	const menustring string = `Quick Menu: Show (a)dmin Key | Adjust (d)atabase | Server (s)ettings | (h)elp`

	const HELP_MESSAGE string = "a:\tshow admin key\nas:\tset admin key\nd:\tdatabase\nh:\thelp\ns:\tserver settings"
	const PRETTY_LINE string = "-------------------------------------------------------------------------------"

	var err error
	var activation_channel chan int = make(chan int, 1)
	var error_channel chan error = make(chan error, 10)
	var uinput string

	go serve(activation_channel, error_channel)
	time.Sleep(time.Millisecond * 300)
	if <-activation_channel != 0 {
		fmt.Printf("server initialization failed\n")
		os.Exit(1)
	}
	println("Server Started")
	for {
		_, err = fmt.Println(menustring)
		Errhandle_Log(err, ERRMSG_WRITE)
		_, err = fmt.Scanf("%s", &uinput)
		println(PRETTY_LINE)
		Errhandle_Log(err, ERRMSG_READ)

		//TODO: Please, at least, make "server settings" not crash the server. It's annoying. REALLY, make the UI usable. A GitHub-interactive website would be cool--really cool. Like, ABRCAMS cool. Let's do it.
		//-Kenton to Kenton
		switch uinput {
		case "a":
			fmt.Printf("%v\n", adminkey)
		case "as":
			UpdateAdminTable(devicedb)
		case "d":
			os.Exit(0)
		case "h":
			println(HELP_MESSAGE)
			println(PRETTY_LINE)
		case "s":
			os.Exit(0)
		}
	}

}
func serve(activation_channel chan int, error_channel chan error) {
	var err error
	var errorchannel chan error
	logflag = flag.Bool("log", true, USTRING_SHOWLOG)
	awayflag = flag.Bool("away", false, USTRING_AWAY)
	flag.Parse()

	_ = InitiateEverything(error_channel)
	log.Printf("channel length = %d", len(activation_channel))
	for len(errorchannel) != 0 {
		err = <-errorchannel
		Errhandle_Log(err, err.Error())
		if err == nil {
			activation_channel <- 0
		} else {
			activation_channel <- 1
		}
	}
	tcpl, err := net.Listen(TCP, IP_main) //change
	Errhandle_Log(err, ERRMSG_NETWORK_CONNECTION)

	///send activation message:
	if err == nil {
		activation_channel <- 0
	} else {
		activation_channel <- 1
	}

	for {
		//fmt.Printf("%sRESTARTING HERE\n%s", ANSIRED, ANSIRESET)
		conn, err := tcpl.Accept()
		Errhandle_Log(err, ERRMSG_NETWORK_CONNECTION)
		if err != nil {
			continue
		}
		go Multiplexer(conn, error_channel)
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
		go Multiplexer(conn, make(chan error))
	}

}
