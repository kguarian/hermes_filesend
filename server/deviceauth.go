//Device authentication code. Monitors validity of new devices and such

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

//creates a file called "devicelists/senders.txt"
func InitDeviceTables() {
	var file *os.File
	var err error

	basedir, err = os.Getwd()
	Errhandle_Exit(err, ERRMSG_GETWD)
	err = os.Chdir(DIR_AUTH)
	Errhandle_Log(err, ERRMSG_FILEIO)
	if err != nil {
		err = os.Mkdir(DIR_AUTH, PERM_RWX_OWNER)
		Errhandle_Exit(err, ERRMSG_FILEIO)
	}
	err = os.Chdir(basedir)
	if err != nil {
		Errhandle_Exit(err, ERRMSG_FILEIO)
	}

	file, err = os.Open(FILE_SENDERLIST)
	Errhandle_Log(err, ERRMSG_FILEIO)
	if os.IsNotExist(err) {
		_, err = os.Create(FILE_SENDERLIST)
		Errhandle_Exit(err, ERRMSG_FILEIO)
		_, err = os.Open(FILE_SENDERLIST)
		Errhandle_Exit(err, ERRMSG_FILEIO)
	}
	file.Close()

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

	Errhandle_Log(err, ERRMSG_FILEIO)
	if err != nil {
		return retdevice, errors.New(ERRMSG_FILEIO)
	}

	mut_devlist.Lock()
	defer mut_devlist.Unlock()

	file, err = os.OpenFile(FILE_SENDERLIST, os.O_RDWR|os.O_APPEND, PERM_RWX_OWNER)
	Errhandle_Log(err, ERRMSG_FILEIO)
	if err != nil {
		return retdevice, errors.New(ERRMSG_FILEIO)
	}
	defer file.Close()

	Errhandle_Log(err, ERRMSG_FILEIO)
	if err != nil {
		return retdevice, errors.New(ERRMSG_FILEIO)
	}

	filewriter = bufio.NewWriter(file)
	//TODO: implement a backup mechanism in case this crashes the program.

	_, err = filewriter.WriteString(string(d.MarshalDevice()) + "\n")
	Errhandle_Log(err, ERRMSG_WRITE)
	if err != nil {
		return retdevice, errors.New(ERRMSG_WRITE)
	}
	err = filewriter.Flush()
	Errhandle_Log(err, ERRMSG_WRITE)
	if err != nil {
		return retdevice, errors.New(ERRMSG_WRITE)
	}
	//housecleaning
	return d, nil
}

//checks devicelists/senders for device entry with the parametrized properties.
func CheckForDevice(userid string, devname string) ([]byte, error) {
	var file *os.File
	var err error
	var filereader *bufio.Reader
	var filereadbuf []byte
	var currdevice device

	if basedir == "" {
		return nil, errors.New(ERRMSG_BASEDIR_NOT_FOUND)
	}

	Errhandle_Log(err, ERRMSG_FILEIO)
	if err != nil {
		return nil, errors.New(ERRMSG_FILEIO)
	}

	mut_devlist.Lock()
	defer mut_devlist.Unlock()

	file, err = os.OpenFile(FILE_SENDERLIST, os.O_RDONLY, PERM_RWX_OWNER)
	Errhandle_Log(err, ERRMSG_FILEIO)
	if err != nil {
		return nil, errors.New(ERRMSG_FILEIO)
	}
	defer file.Close()

	Errhandle_Log(err, ERRMSG_FILEIO)
	if err != nil {
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
			return filereadbuf, nil
		}
	}
	return nil, nil
}
