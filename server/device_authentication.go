//Device authentication code. Monitors validity of new devices and such

package main

import (
	"bufio"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
)

const (
	PERM_RWX_OWNER = 0b1_111_000_000
)

var basedir string

//creates a file called "devicelists/senders.txt"
func InitDeviceTables() *sql.DB {
	Info_Log("Initializing Device Database")
	var dbfile *os.File
	var err error
	var mysqlpw_byte []byte
	var mysqlpw string
	//var mysqlpw string

	devdb_mutex.Lock()
	defer devdb_mutex.Unlock()

	dbfile, err = os.Open(DEVICE_DB_PATH)
	if err != nil {
		fmt.Printf("%s\n", err)
		if _, ok := err.(*os.PathError); ok {
			dbfile, err = os.Create(DEVICE_DB_PATH)
			if err != nil {
				log.Panicf("unknown IO error in InitDeviceTables(), case 1\n")
			}
		} else {
			log.Panicf("unknown IO error in InitDeviceTables(), case 2\n")
		}
	}
	defer dbfile.Close()

	f, err := os.Open("../../../../misc/mysqlpw/mysqlpw.txt")
	Errhandle_Exit(err, ERRMSG_FILEIO)
	defer f.Close()
	mysqlpw_byte, _, _ = bufio.NewReader(f).ReadLine()
	mysqlpw = string(mysqlpw_byte)
	Info_Log("password: " + mysqlpw)
	mysqldb, err := sql.Open("mysql", fmt.Sprintf("root:%s@/hermes", mysqlpw)) // Open the created SQLite File
	if err != nil {
		log.Fatalf("here: %s\n", err)
	}
	// Info_Log(mysqlpw)

	err = DB_CreateDeviceTable(mysqldb)
	if err != nil {
		Errhandle_Exit(err, ERRMSG_DBEXISTS)
	}

	devicedb = mysqldb
	Info_Log("Finished Initializing Device Database")
	return mysqldb
}

//check devices/senders.txt for device entry with matching userid and devicename
func AddDevice(d Device) error {
	var err error
	var user User
	var currUser User

	mut_devlist.Lock()
	defer mut_devlist.Unlock()

	user, err = DB_GetUser(devicedb, d.Username)
	Errhandle_Log(err, ERRMSG_DEVICENOTFOUND)
	if err != nil {
		new_uuid, err := uuid.NewUUID()
		Errhandle_Log(err, ERRMSG_CREATE_UUID)
		if err != nil {
			return err
		}
		devlist := make(map[uuid.UUID]Device)
		devlist[d.Device_uuid] = d
		user = User{Username: d.Username, User_uuid: new_uuid, Devicelist: devlist}
		err = DB_AddUser(devicedb, d.Username, user)
		Errhandle_Log(err, ERRMSG_DB_ATTEMPTED_INSERT_DUPLICATE)
		return nil
	}
	currUser, err = DB_GetUser(devicedb, d.Username)
	Errhandle_Log(err, ERRMSG_DEVICENOTFOUND)
	if err != nil {
		return err
	}
	currUser.Devicelist[d.Device_uuid] = d
	err = DB_AddUser(devicedb, d.Username, currUser)
	return err
}

//checks devicelists/senders for device entry with the parametrized properties.
func CheckForDevice(userid string, devname string) (Device, error) {
	var devslice []Device
	var err error
	var retDevice Device

	devslice, err = DB_GetDeviceSlice(devicedb, userid)
	Info_Log(devslice)
	Errhandle_Log(err, ERRMSG_DB_SELECT)
	if err != nil {
		return retDevice, err
	}
	if len(devslice) == 0 {
		return retDevice, errors.New(ERRMSG_DEVICENOTFOUND)
	} else {
		for _, device := range devslice {
			if device.Devicename == devname {
				return device, nil
			}
		}
		return retDevice, errors.New(ERRMSG_DEVICENOTFOUND)
	}
}
