package main

//USING BIG-ENDIAN
//REF: 	https://www.ibm.com/docs/en/error?originalUrl=SSB27U_6.4.0/com.ibm.zvm.v640.kiml0/asonetw.htm#:~:text=The%20network%20byte%20order%20is,confusion%20because%20of%20byte%20ordering.
import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
)

type Netmessage struct {
	Message string
}

//SendStruct is used to send **JSON-MARSHALLED structs.
func SendStruct(i interface{}, c net.Conn) error {
	const length_of_int = 4
	var b []byte
	var err error
	var length uint32
	var recvlength int
	var content_lenbuf []byte
	var client_lenbuf []byte
	var errbuf []byte

	if i == nil {
		return errors.New(ERRMSG_NILPTR)
	}
	b, err = json.Marshal(i)

	Errhandle_Log(err, ERRMSG_JSON_MARSHALL)
	if err != nil {
		return err
	}
	content_lenbuf = make([]byte, length_of_int)
	client_lenbuf = make([]byte, length_of_int)
	errbuf = make([]byte, 1)

	length = uint32(len(b))
	binary.BigEndian.PutUint32(content_lenbuf, length)

	//COMPLIMENTARY: Send length of JSON object to send
	recvlength, err = c.Write(content_lenbuf)
	Errhandle_Log(err, ERRMSG_NETWORK_WRITE)
	if recvlength != length_of_int {
		return errors.New(ERRMSG_NETWORK_WRITE)
	}
	//COMPLIMENTARY: Receive same length as confirmation of receipt
	c.Read(client_lenbuf)
	if bytes.Equal(content_lenbuf, client_lenbuf) {
		//COMPLIMENTARY: Send json struct
		_, err = c.Write(b)
		return err
	} else {
		//COMPLIMENTARY: Send error message
		errbuf[0] = NETCODE_ERR
		c.Write(errbuf)
		return errors.New(ERRMSG_NETWORK_WRITE)
	}
}

//PASS A POINTER TO RECEIVE THE STRUCT
func ReceiveStruct(i interface{}, c net.Conn) error {
	const length_of_int = 4
	var intbuf []byte = make([]byte, length_of_int)
	var contentlength int
	var contentbuf []byte
	var recvlength int
	var err error

	//COMPLIMENTARY: Receive length of JSON object to make buffer
	recvlength, err = c.Read(intbuf)
	Errhandle_Log(err, ERRMSG_NETWORK_READ)
	if recvlength != length_of_int {
		return errors.New(ERRMSG_NETWORK_READ)
	}
	contentlength = int(binary.BigEndian.Uint32(intbuf))
	contentbuf = make([]byte, contentlength)
	//COMPLIMENTARY: Send received length as confirmation of receipt
	_, err = c.Write(intbuf)
	Errhandle_Log(err, ERRMSG_NETWORK_WRITE)
	//COMPLIMENTARY: Receive either JSON struct or error message
	recvlength, err = c.Read(contentbuf)
	Errhandle_Log(err, ERRMSG_NETWORK_READ)
	//error message or bad transmission
	if recvlength != contentlength {
		return errors.New(ERRMSG_NETWORK_READ)
	}
	//should work unless wrong struct passed in
	err = json.Unmarshal(contentbuf, i)
	return err
}

func NewNetmessage(message string) Netmessage {
	return Netmessage{message}
}

func PassRequest(msg *Netmessage, send interface{}, recv interface{}) error {

	var connServer net.Conn
	var err error
	var req http.Request

	connServer, err = net.Dial(TCP, IP_SERVER)
	req = http.Request{URL: &url.URL{}, Method: http.MethodGet, Header: map[string][]string{"Nativegoclient": {"true"}}}

	Errhandle_Log(err, ERRMSG_NETWORK_DIAL)
	if err != nil {
		return err
	}
	defer connServer.Close()
	err = req.Write(connServer)
	Errhandle_Log(err, ERRMSG_NETWORK_SEND_STRUCT)
	if err != nil {
		return err
	}
	err = SendStruct(msg, connServer)
	Errhandle_Log(err, ERRMSG_NETWORK_SEND_STRUCT)
	if err != nil {
		return err
	}
	err = ReceiveStruct(msg, connServer)
	Errhandle_Log(err, ERRMSG_NETWORK_RECEIVE_STRUCT)
	if err != nil {
		return err
	}
	err = SendStruct(send, connServer)
	Errhandle_Log(err, ERRMSG_NETWORK_SEND_STRUCT)
	if err != nil {
		return err
	}
	err = ReceiveStruct(recv, connServer)
	Errhandle_Log(err, ERRMSG_NETWORK_READ)
	if err != nil {
		return err
	}
	fmt.Printf("response: %v\n", recv)
	return nil
}
