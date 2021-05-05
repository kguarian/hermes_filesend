package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"
)

var db_mutex sync.Mutex = sync.Mutex{}
var devicedb *sql.DB

//NOTE: user_devices is name of table
func DB_CreateDeviceTable(db *sql.DB) error {
	createURLTableSQL := `CREATE TABLE user_devices (
		"username"			TEXT,
		"device_slice"		TEXT
	  );` // SQL Statement for Create Table

	//fmt.Println("Create url_status table...")
	statement, err := db.Prepare(createURLTableSQL) // Prepare SQL Statement
	if err != nil {
		return err
	}
	defer statement.Close()
	statement.Exec() // Execute SQL Statements
	fmt.Println("url_status table created")
	return nil
}

func DB_InsertDeviceSlice(db *sql.DB, username string, deviceslice []Device) error {
	//fmt.Println("Inserting URL Status ...")
	insertUrlStatusSQL := `INSERT INTO user_devices(username, device_slice) VALUES (?,?)`
	db_mutex.Lock()
	defer db_mutex.Unlock()
	statement, err := db.Prepare(insertUrlStatusSQL) // Prepare statement.
	// This is good to avoid SQL injections
	if err != nil {
		return err
	}
	//timestamp := time.Now()
	//fmt.Printf("%s INSERT %s  %3d  %t   %s   %s   %s\n", digest_num, digest_num, line_in_digest, valid, timestamp, orig_url, derived_url)
	jsonstring, err := json.Marshal(&deviceslice)
	if err != nil {
		log.Printf("Marshal Error.\n")
	}
	_, err = statement.Exec(username, jsonstring)
	db_mutex.Unlock()
	if err != nil {
		log.Printf("Execute INSERT INTO failed: %s\n", err.Error())
		time.Sleep(500 * time.Second)
	}
	statement.Close()
	return nil
}

//err being nil does not imply that the returned device slice is not nil.
//err being non-nil implies that you should not use the returned slice.
func DB_GetDeviceSlice(db *sql.DB, username string) ([]Device, error) {
	var err error
	var sqlrows *sql.Rows
	var jsondevslice string
	var retslice []Device
	//fmt.Println("Inserting URL Status ...")
	DB_selectdevslice := `SELECT device_slice from user_devices WHERE username == ?`
	db_mutex.Lock()
	defer db_mutex.Unlock()
	sqlrows, err = db.Query(DB_selectdevslice, username) // Prepare statement.
	// This is good to avoid SQL injections
	if err != nil {
		return nil, err
	}
	defer sqlrows.Close()
	if sqlrows.Next() {
		sqlrows.Scan(&jsondevslice)
	} else {
		retslice = nil
	}
	//fmt.Printf("SELECTed entry: %s\n", jsondevslice)
	if len(jsondevslice) == 0 {
		return nil, nil
	}
	err = json.Unmarshal([]byte(jsondevslice), &retslice)
	if err != nil {
		log.Printf("failed unmarshal\n")
	}
	return retslice, err
}
