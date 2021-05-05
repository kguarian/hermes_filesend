package main

import (
	"encoding/json"
	"fmt"
	"log"

	"syscall/js"
)

func main() {
	c := make(chan struct{})
	js.Global().Set("DeviceConn", js.FuncOf(DeviceConn))
	<-c
}

//val[0]=userid
//val[1]=devicename
func DeviceConn(this js.Value, val []js.Value) interface{} {
	var userid, devicename string = fmt.Sprintf("%s", val[0]), fmt.Sprintf("%s", val[1])

	var devinf deviceinfo
	var received_device device
	//var connServer net.Conn
	var err error
	var b []byte
	var response string

	ws := js.Global().Get("WebSocket").New("ws://" + IP_SERVER)
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
		return nil
	}))

	ws.Call("addEventListener", "message", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		response = fmt.Sprintf("%s", args[0])
		log.Printf("%s\n", response)
		if response != NETREQ_NEWDEVICE_JAVASCRIPT {
			log.Printf("Unexpected response from server: %s\n Expected: %s\n",
				response, NETREQ_NEWDEVICE_JAVASCRIPT)
			return nil
		}
		devinf = deviceinfo{Userid: userid, Devicename: devicename}
		b, err = json.Marshal(&devinf)

		if err != nil {
			Errhandle_Log(err, ERRMSG_JSON_MARSHALL)
		}
		ws.Call("send", b)
		return nil
	}))

	ws.Call("addEventListener", "message", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		response = fmt.Sprintf(response, args[0])
		log.Printf("received device: %s\n", response)
		return nil
	}))

	fmt.Printf("it's this: %s\n", response)
	err = json.Unmarshal([]byte(response), received_device)
	return response
}
