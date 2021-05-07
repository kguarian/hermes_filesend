package main

import (
	"bufio"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
)

/*Dr. Muller:
Have your Ambassador function take the first message and determine what kind of client you're dealing with.

Method: Capture transient first-message. If something expected, then handle as expected, otherwise, flag it as an exception, log it, leave.

Note: If you need to add a preamble message to a protocol so that we know who we're talking to, then it's not bloat. Just do it.

If you're trying to put a message in a byte string, then try to read the whole message in [size of buffer] bytes. If you don't have all the bytes in the message,
then read the rest of the buffer again.

must determine when to stop reading a message. Must know length or terminal character or something.

General Strategy (Covered in Stevens): how to encapsulate message in byte string
*/

//handles requests for the main
func Ambassador(conn net.Conn) {
	var msg Netmessage
	var err error
	var clienttype int //enum (const)
	clienttype, err = ClientIdentifier(conn)
	if err != nil {
		Errhandle_Log(err, "Error occured in ClientIdentifier function. Terminating connection./n")
		return
	}
	fmt.Fprintf(os.Stderr, "Connection received from device with enum: %d\n", clienttype)

	if clienttype == nativegoclient {
		err = ReceiveStruct(&msg, conn)
		Errhandle_Log(err, ERRMSG_NETWORK_RECEIVE_STRUCT)
		if err != nil {
			return
		}
	} else if clienttype == jsclient {
		err = ReceiveStruct_JSClient(&msg, conn)
		if err != nil {
			return
		}
	}
	fmt.Printf("MESSAGE RECEIVED: %v\n", msg)
	switch msg.Message {
	case NETREQ_NEWDEVICE:
		RequestDevice(conn)
	case NETREQ_NEWDEVICE_JAVASCRIPT:
		RequestDevice_Websocket(conn)
	default:
		conn.Close()
	}
}

func ClientIdentifier(conn net.Conn) (int, error) {
	var httpreq *http.Request
	var err error
	var bufioReader *bufio.Reader = bufio.NewReader(conn)
	var retval int
	if err != nil {
		return retval, err
	}
	httpreq, err = http.ReadRequest(bufioReader)
	Errhandle_Log(err, ERRMSG_NETWORK_PARSE_HTTPREQ)
	if err != nil {
		return -1, err
	}
	bufioReader = bufio.NewReader(httpreq.Body)
	fmt.Printf("Header: %v\n", httpreq.Header)
	if httpreq.Header["Sec-Websocket-Version"] != nil {
		var clienthash string = httpreq.Header["Sec-Websocket-Key"][0]
		clienthash += "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"

		//https: //yalantis.com/blog/how-to-build-websockets-in-go/
		hash := sha1.New()
		hash.Write([]byte(clienthash))
		var rethash string = base64.StdEncoding.EncodeToString(hash.Sum(nil))
		retstring := "HTTP/1.1 101 Switching Protocols\r\nUpgrade: websocket\r\nConnection: Upgrade\r\nSec-Websocket-Accept: %s\r\n\r\n"
		retstring = fmt.Sprintf(retstring, rethash)
		fmt.Printf("\n\rresponse:\n%s\n", retstring)
		conn.Write([]byte(retstring))
		return jsclient, nil
	} else if httpreq.Header["Nativegoclient"] != nil {
		return nativegoclient, nil
	} else {
		return -1, nil
	}
}

//Request to add/register device.
func RequestDevice(conn net.Conn) (Device, error) {
	var retdevice Device
	var ipstring string
	var ip net.IP
	var cdinf DeviceInfo
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
	log.Printf("native go client: userid: %s, device: %s", cdinf.Userid, cdinf.Devicename)
	ipstring = strings.Split(conn.RemoteAddr().String(), ":")[0]
	ip = net.ParseIP(ipstring)
	retdevice, err = NewDevice(cdinf.Userid, cdinf.Devicename, ip)
	Errhandle_Log(err, ERRMSG_CREATE_DEVICE)
	err = SendStruct(&retdevice, conn)
	Errhandle_Log(err, ERRMSG_NETWORK_SEND_STRUCT)
	//return or return, because we can't return a nil device
	return retdevice, err
}

func RequestDevice_Websocket(conn net.Conn) (Device, error) {
	var retdevice Device
	var err error
	var cdinf DeviceInfo
	var ip net.IP
	var ipstring string

	err = ReceiveStruct_JSClient(&cdinf, conn)
	if err != nil {
		return retdevice, err
	}
	ipstring = strings.Split(conn.RemoteAddr().String(), ":")[0]
	ip = net.ParseIP(ipstring)
	retdevice, err = NewDevice(cdinf.Userid, cdinf.Devicename, ip)
	Errhandle_Log(err, ERRMSG_CREATE_DEVICE)
	log.Printf("DEVICE: %v\n", retdevice)
	err = SendStruct_JSClient(&retdevice, conn)
	Errhandle_Log(err, ERRMSG_NETWORK_SEND_STRUCT)
	//return or return, because we can't return a nil device
	return retdevice, err
}
