package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"os"
)

const (
	PERM_RWX_OWNER = 0b1_111_000_000
)

var basedir string

func InitDeviceTables() {
	var err error

	basedir, err = os.Getwd()
	Errhandle_Exit(err, ERRMSG_GETWD)
	err = os.Chdir(DIR_AUTH)
	Errhandle_Log(err, ERRMSG_FILEIO)
	if err != nil {
		err = os.Mkdir(DIR_AUTH, PERM_RWX_OWNER)
		Errhandle_Exit(err, ERRMSG_FILEIO)
		err = os.Chdir(DIR_AUTH)
		Errhandle_Exit(err, ERRMSG_FILEIO)
	}

	_, err = os.Open(FILE_SENDERLIST)
	Errhandle_Log(err, ERRMSG_FILEIO)
	if os.IsNotExist(err) {
		_, err = os.Create(FILE_SENDERLIST)
		Errhandle_Exit(err, ERRMSG_FILEIO)
		_, err = os.Open(FILE_SENDERLIST)
		Errhandle_Exit(err, ERRMSG_FILEIO)
	}

	//housecleaning
	os.Chdir(basedir)
}

//check devices/senders.txt for device entry with matching userid and devicename
func AddDevice(d device) (device, error) {
	var retdevice device
	var file *os.File
	var filewriter *bufio.Writer
	var err error

	if basedir == "" {
		return retdevice, errors.New(ERRMSG_BASEDIR_NOT_FOUND)
	}

	err = os.Chdir(DIR_AUTH)
	Errhandle_Log(err, ERRMSG_FILEIO)
	if err != nil {
		goHome()
		return retdevice, errors.New(ERRMSG_FILEIO)
	}

	file, err = os.OpenFile(FILE_SENDERLIST, os.O_RDWR|os.O_APPEND, PERM_RWX_OWNER)
	Errhandle_Log(err, ERRMSG_FILEIO)
	if err != nil {
		goHome()
		return retdevice, errors.New(ERRMSG_FILEIO)
	}
	defer file.Close()

	Errhandle_Log(err, ERRMSG_FILEIO)
	if err != nil {
		goHome()
		return retdevice, errors.New(ERRMSG_FILEIO)
	}

	filewriter = bufio.NewWriter(file)
	//TODO: implement a backup mechanism in case this crashes the program.
	//note: 6/7/20: probably never happening. This is Go, not C.

	_, err = filewriter.WriteString(string(d.MarshalDevice()) + "\n")
	Errhandle_Log(err, ERRMSG_WRITE)
	if err != nil {
		goHome()
		return retdevice, errors.New(ERRMSG_WRITE)
	}
	err = filewriter.Flush()
	Errhandle_Log(err, ERRMSG_WRITE)
	if err != nil {
		goHome()
		return retdevice, errors.New(ERRMSG_WRITE)
	}
	//housecleaning
	goHome()
	return d, nil
}

func CheckForDevice(userid string, devname string) ([]byte, error) {
	var file *os.File
	var err error
	var filereader *bufio.Reader
	var filereadbuf []byte
	var currdevice device

	if basedir == "" {
		return nil, errors.New(ERRMSG_BASEDIR_NOT_FOUND)
	}

	err = os.Chdir(DIR_AUTH)
	Errhandle_Log(err, ERRMSG_FILEIO)
	if err != nil {
		goHome()
		return nil, errors.New(ERRMSG_FILEIO)
	}
	file, err = os.OpenFile(FILE_SENDERLIST, os.O_RDONLY, PERM_RWX_OWNER)
	Errhandle_Log(err, ERRMSG_FILEIO)
	if err != nil {
		goHome()
		return nil, errors.New(ERRMSG_FILEIO)
	}
	defer file.Close()

	Errhandle_Log(err, ERRMSG_FILEIO)
	if err != nil {
		goHome()
		return nil, errors.New(ERRMSG_FILEIO)
	}
	filereader = bufio.NewReader(file)

	for {
		filereadbuf, err = filereader.ReadBytes('\n')
		Errhandle_Log(err, ERRMSG_READ)
		if err != nil {
			break
		}
		//Sometimes, newlines are added to the file, so until we fix that, we'll just not consider them.
		//TODO: That was a patch, not a fix
		if len(filereadbuf) < 2 {
			continue
		}
		err = json.Unmarshal(filereadbuf, &currdevice)
		Errhandle_Log(err, ERRMSG_JSON_UNMARSHALL)
		if currdevice.Userid == userid && currdevice.Devicename == devname {
			goHome()
			return filereadbuf, nil
		}
	}
	goHome()
	return nil, nil
}

func goHome() {
	err := os.Chdir(basedir)
	Errhandle_Exit(err, ERRMSG_FILEIO)
}
