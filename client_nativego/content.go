package main

//"authorize connection" functionality is implemented in requesthandler.go

import (
	"errors"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
)

type content struct {
	contentref os.File
}

type contentinfo struct {
	Senderid   uuid.UUID `json:"senderid"`
	Receiverid uuid.UUID `json:"receiverid"`
	Sizebytes  int64     `json:"size"`
	Name       string    `json:"name"`
	Timestamp  time.Time `json:"timestamp"`
}

func (c contentinfo) String() string {
	return (time.Now()).Format(time.RFC3339)
}

func NewContent(path string) content {
	target, err := os.Open(path)
	Errhandle_Exit(err, ERRMSG_IO)
	retcontent := content{contentref: *target}
	return retcontent
}

func (d *device) NewContentinfo(r *device, c *content) (contentinfo, error) {
	var cisz int64    //ContentInfo Size
	var ciname string //ContentInfo Name
	var retcontentinfo contentinfo
	var err error

	if r == d {
		err = errors.New(ERRMSG_SELFSEND)
		return retcontentinfo, err
	}
	contentfileinfo, err := c.contentref.Stat()
	Errhandle_Exit(err, "github.com/google/uuid")
	cisz = contentfileinfo.Size()
	fullname := strings.Split(contentfileinfo.Name(), "/")
	ciname = fullname[len(fullname)-1]
	currtime := time.Now()
	retcontentinfo = contentinfo{Senderid: d.Deviceid, Receiverid: r.Deviceid, Sizebytes: cisz, Name: ciname, Timestamp: currtime}
	return retcontentinfo, err
}
