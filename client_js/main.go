package main

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"syscall/js"
)

var wgmain sync.WaitGroup

func main() {
	c := make(chan bool)
	js.Global().Set("DeviceConn", js.FuncOf(DeviceConn))
	<-c
}

//val[0]=userid
//val[1]=devicename
func DeviceConn(this js.Value, val []js.Value) interface{} {

	var devinf deviceinfo
	var received_device device

	//both Uint8Arrays
	var userid_jsval js.Value = val[0]
	var devicename_jsval js.Value = val[1]
	//create strings
	var userid string = userid_jsval.String()
	var devicename string = devicename_jsval.String()
	log.Printf("userid: %v, devicename: %v\n", userid_jsval.Type(), devicename_jsval.Type())

	//var connServer net.Conn
	var err error
	var b []byte
	var retstring string

	//because this function exits before network processes finish:

	ws := js.Global().Get("WebSocket").New("ws://" + IP_SERVER)
	ws.Set("binaryType", "arraybuffer")
	ws.Call("addEventListener", "open", js.FuncOf(func(this js.Value, args []js.Value) interface{} {

		protocol := NewNetmessage(NETREQ_NEWDEVICE_JAVASCRIPT)
		b, err = json.Marshal(&protocol)
		if err != nil {
			log.Printf("%s\n", err)
			return nil
		}
		jsvalue := js.Global().Get("Uint8Array").New(len(b))
		js.CopyBytesToJS(jsvalue, b)
		ws.Call("send", jsvalue)

		devinf = deviceinfo{Userid: userid, Devicename: devicename}
		b, err = json.Marshal(&devinf)

		if err != nil {
			Errhandle_Log(err, ERRMSG_JSON_MARSHALL)
		}
		jsvalue = js.Global().Get("Uint8Array").New(len(b))
		js.CopyBytesToJS(jsvalue, b)
		ws.Call("send", jsvalue)
		return nil
	}))

	ws.Call("addEventListener", "message", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		jsvalue := js.Global().Get("Uint8Array").New(args[0].Get("data"))
		jslength := args[0].Get("data").Get("byteLength").Int()

		b = make([]byte, jslength)
		js.CopyBytesToGo(b, jsvalue)

		err = json.Unmarshal(b, &received_device)

		if err != nil {
			Errhandle_Log(err, ERRMSG_JSON_MARSHALL)
		}

		log.Printf("received device: %v", received_device)
		js.Global().Get("document").Call("getElementById", "deviceid").Set("innerHTML", js.ValueOf(fmt.Sprintf("%v", received_device)))
		return nil
	}))

	return js.ValueOf(retstring)
}
