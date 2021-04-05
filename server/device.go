package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"

	"github.com/google/uuid"
)

//TODO: REVISE, consider character limit for device names and usernames.
var INVALIDUSERCHARS []rune = []rune{'\t', ' ', ':', '[', ']', '.'}

//TODO: COMPLETE THE STATEMENT BELOW AND IMPLEMENT
//const namelengthlimit int = [decide on a limit and put it here]

//for internal use, mainly. This is shared with the client as of 4/4/2021
type device struct {
	Userid      string                 `json:"userid"`
	Devicename  string                 `json:"devname"`
	Deviceid    uuid.UUID              `json:"devid"`
	Ipaddr      net.IP                 `json:"ipaddr"`
	Indata      chan byte              `json:"-"`
	Inrequests  chan contentinfo       `json:"-"`
	Outdata     chan byte              `json:"-"`
	Consentlist map[string]contentinfo `json:"clist"`
	Online      bool                   `json:"-"`
}

//Sent from client to server on initiated contact.
type deviceinfo struct {
	Userid     string `json:"userid"`
	Devicename string `json:"devname"`
}

//Constructor
//NewDevices creates a new device, but returns an error iff the parametrized id is invalid
func NewDevice(userid string, devicename string, ipaddr net.IP) (device, error) {
	const invalidusername string = "invalid username string"
	var retdevice device
	var err error
	var devicejson []byte
	if !EvalName(userid) {
		err = errors.New(invalidusername)
		return retdevice, err
	}
	devicejson, err = CheckForDevice(userid, devicename)
	Errhandle_Log(err, ERRMSG_DEVICECHECK)

	if err != nil || devicejson == nil {
		cl := make(map[string]contentinfo)
		deviceid := uuid.New()
		retdevice = device{Userid: userid, Deviceid: deviceid, Devicename: devicename, Ipaddr: ipaddr, Consentlist: cl, Online: false}
		AddDevice(retdevice)
	} else {
		err = json.Unmarshal(devicejson, &retdevice)
		Errhandle_Log(err, ERRMSG_JSON_UNMARSHALL)
		if err != nil {
			return retdevice, errors.New(ERRMSG_JSON_UNMARSHALL)
		}
	}

	in := make(chan byte, 10*1024)
	inreq := make(chan contentinfo, 10)
	out := make(chan byte, 10)

	retdevice.Indata = in
	retdevice.Outdata = out
	retdevice.Inrequests = inreq
	return retdevice, nil
}

//This is an incredibly useless wrapper function that harms no one.
func (d *device) MarshalDevice() []byte {
	retinfo, err := json.Marshal(d)
	Errhandle_Log(err, ERRMSG_JSON_MARSHALL)
	if err != nil {
		return nil
	}
	return retinfo
}

//TODO: Actually implement a login protocol
func (d *device) Login() {
	d.Online = true
}

//TODO: Implement a login/logout system.
func (d *device) Logout() {
	d.Online = false
}

//TODO: Implement a networked form of this.
//RequestConsent sends a single-file consent-to-transfer request from one device to another
func (d *device) RequestConsent(recipientdevice device, c contentinfo) error {
	//pointer not nil; checked above
	if !recipientdevice.Online {
		SetConsoleColor(RED)
		fmt.Printf("device [%s] is offline. Request Canceled.\n", recipientdevice.Deviceid)
		SetConsoleColor(RESET)
		return errors.New(ERRMSG_DEVICEOFFLINE)
	}
	recipientdevice.Inrequests <- c
	return nil
}

//TODO: Networked form
func (d *device) EvalConsent(sender *device) error {
	var character byte
	var cont contentinfo
	var err error
	var ok bool

	for len(d.Inrequests) > 0 {
		cont, ok = <-d.Inrequests
		fmt.Printf("%s channel depth: %d\n", d.MarshalDevice(), len(d.Inrequests))
		if !ok {
			err = errors.New(ERRMSG_CHANNEL_OPERATION)
			return err
		}
		if cont.Senderid == sender.Deviceid {
			fmt.Printf("%s requests to send file [%s], size: %d bytes.\n", cont.Senderid, cont.Name, cont.Sizebytes)
			fmt.Printf("Approve? (Y/*): ")
			fmt.Scanf("%c\n", &character)
			fmt.Println(character == 'Y')
			d.Consentlist[cont.Name] = cont
		}
	}
	return nil
}

//TODO: Networked form
func Approveconsent(authtoken byte, c *contentinfo) (string, error) {
	if authtoken == 'Y' {
		return c.Senderid.String() + ":" + c.Name + ":" + string(c.Sizebytes), nil
	}
	return "error", errors.New(ERRMSG_CONSENT_TO_SEND)
}

//TODO: implement
func (d *device) SendContent(c *contentinfo) {

}

func EvalName(name string) bool {
	for _, c := range name {
		for _, c2 := range INVALIDUSERCHARS {
			if c == c2 {
				return false
			}
		}
	}
	return true
}
