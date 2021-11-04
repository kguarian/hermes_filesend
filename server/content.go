//This file should contain everything that has to do with the representation of files transmitted by the service.

package main

//"authorize connection" functionality is implemented in requesthandler.go

import (
	"errors"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
)

//represents file for sending.
//TODO: consider representing files with a map[int][]byte
type content struct {
	contentref os.File
}

//for machine-machine communication about files that will be shared.
type contentinfo struct {
	Senderid   uuid.UUID `json:"senderid"`
	Receiverid uuid.UUID `json:"receiverid"`
	Sizebytes  int64     `json:"size"`
	Name       string    `json:"name"`
	Timestamp  time.Time `json:"timestamp"`
}

//Constructor
func NewContent(path string) content {
	target, err := os.Open(path)
	Errhandle_Exit(err, ERRMSG_IO)
	retcontent := content{contentref: *target}
	return retcontent
}

//Constructor
func (d *Device) NewContentinfo(r *Device, c *content) (contentinfo, error) {
	var cisz int64    //ContentInfo Size
	var ciname string //ContentInfo Name
	var retcontentinfo contentinfo
	var err error

	if r == d {
		err = errors.New(ERRMSG_SELFSEND)
		return retcontentinfo, err
	}
	contentfileinfo, err := c.contentref.Stat()
	Errhandle_Exit(err, ERRMSG_IO)
	cisz = contentfileinfo.Size()
	fullname := strings.Split(contentfileinfo.Name(), "/")
	ciname = fullname[len(fullname)-1]
	currtime := time.Now()
	retcontentinfo = contentinfo{Senderid: d.Device_uuid, Receiverid: r.Device_uuid, Sizebytes: cisz, Name: ciname, Timestamp: currtime}
	return retcontentinfo, err
}
