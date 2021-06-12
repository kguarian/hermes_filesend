package main

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync"
	"syscall/js"
	"time"
	"unicode"
)

var wgmain sync.WaitGroup

func main() {
	c := make(chan bool)
	js.Global().Set("DeviceConn", js.FuncOf(DeviceConn))
	js.Global().Set("door_encode", js.FuncOf(timestamper))
	js.Global().Set("door_decode", js.FuncOf(timereader))
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

//works for doors of
func timestamper(this js.Value, val []js.Value) interface{} {
	var unixtime int64
	var input []byte
	var timeval []byte
	var hashval []byte
	var done bool
	var retval js.Value

	for done = false; !done; {
		unixtime = int64(time.Now().UnixNano())
		input = []byte(val[0].String())
		timeval = make([]byte, 8)
		binary.LittleEndian.PutUint64(timeval, uint64(unixtime))
		hashval = xor(timeval, input)
		for i, v := range hashval {
			if !unicode.IsGraphic(rune(v)) || v == ' ' {
				break
			}
			if i == len(hashval)-1 {
				done = true
				Info_Log("Found good combo")
			}
		}
		unixtime = time.Now().UnixNano()
	}

	retval = js.ValueOf(string(hashval))
	Info_Log(retval.String() == string(hashval))

	var s1, s2 string

	s1 = string(hashval)
	s2 = js.ValueOf(string(hashval)).String()

	fmt.Printf("comparison between %s and %s: %d\n", s1, s2, strings.Compare(s1, s2))

	for i := 0; i < len(s1); i++ {
		fmt.Printf("%8b ", s1[i])
	}
	println()
	for i := 0; i < len(s2); i++ {
		fmt.Printf("%8b ", s2[i])
	}
	println()

	return retval
}

/*
	Workflow:
	params: val[0] = key to hash
	return: hash

	How:
	1) make 64bit unix time (big endian for reading and easy bitshifts)
	2) XOR each time byte with a key byte if it exists, otherwise just toss the time in.
	hash=8-byte result of this operation.

	Reverse workflow:

	params: key, hash
	result: original time

	1) XOR each hash byte with a key byte if it exists, otherwise just take byte of hash as time.
	2) Make 64bit unix time
*/
func timereader(this js.Value, val []js.Value) interface{} {
	var unixtime uint64
	var item string
	var hash string
	var itemslice []byte
	var hashslice []byte
	var err error

	Info_Log(val[0].String())
	_, err = fmt.Sscanf(val[0].String(), "[ %s %s ]", &item, &hash)
	if err != nil {
		Errhandle_Exit(err, "scanning values into string")
	}
	hashslice = []byte(hash)
	itemslice = []byte(item)

	unixtime = binary.LittleEndian.Uint64(xor(hashslice, itemslice))

	fmt.Printf("%64b\n", unixtime)

	return js.ValueOf(time.Unix(0, int64(unixtime)).Format(time.ANSIC))
}

/*
	Description:
	if a and b are of equal length, xor will return a slice where
		retslice[i] = a[i]^b[i] for i < len(a)

	if a and b are not of equal size, suppose len(a) < len(b).
		xor will let retslice[i] = a[i] ^ b[i] if i < len(a)
		xor will let retslice[i] = 0 if i > len(a)
*/
func xor(a, b []byte) (retslice []byte) {
	if len(a) == len(b) {
		retslice = make([]byte, len(a))
		for i := 0; i < len(a); i++ {
			retslice[i] = a[i] ^ b[i]
		}
	} else {
		if len(a) > len(b) {
			var i int
			retslice = make([]byte, len(a))
			for i = 0; i < len(b); i++ {

				retslice[i] = a[i] ^ b[i]
			}
			for ; i < len(a); i++ {

				retslice[i] = a[i]
			}
		} else {
			retslice = make([]byte, len(b))
			var i int
			for i = 0; i < len(a); i++ {

				retslice[i] = a[i] ^ b[i]
			}
			for ; i < len(b); i++ {

				retslice[i] = b[i]
			}
		}
	}
	return
}
