//Device authentication code. Monitors validity of new devices and such

package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

const (
	PERM_RWX_OWNER = 0b1_111_000_000
)

var basedir string

//creates a file called "devicelists/senders.txt"
func InitDeviceTables() *sql.DB {
	var dbfile *os.File
	var err error

	db_mutex.Lock()
	defer db_mutex.Unlock()

	dbfile, err = os.Open(DEVICE_DB_PATH)
	if err != nil {
		fmt.Printf("%s", err)
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

	sqliteDatabase, err := sql.Open("sqlite3", DEVICE_DB_PATH) // Open the created SQLite File
	if err != nil {
		log.Fatalf("%s\n", err)
	}

	err = DB_CreateDeviceTable(sqliteDatabase)
	if err != nil {
		Errhandle_Log(err, ERRMSG_DBEXISTS)
	}

	devicedb = sqliteDatabase
	return sqliteDatabase
}

//check devices/senders.txt for device entry with matching userid and devicename
func AddDevice(d Device) error {
	var err error
	var devslice []Device

	mut_devlist.Lock()
	defer mut_devlist.Unlock()

	devslice, err = DB_GetDeviceSlice(devicedb, d.Userid)
	if err != nil {
		return err
	}
	if len(devslice) == 0 {
		devslice = make([]Device, 1)
		devslice[0] = d
		DB_InsertDeviceSlice(devicedb, d.Userid, devslice)
		return nil
	} else {
		devslice = append(devslice, d)
		DB_InsertDeviceSlice(devicedb, d.Userid, devslice)
		return nil
	}
}

//checks devicelists/senders for device entry with the parametrized properties.
func CheckForDevice(userid string, devname string) (Device, error) {
	var devslice []Device
	var err error
	var retDevice Device

	devslice, err = DB_GetDeviceSlice(devicedb, userid)
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
