package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"hermes/server"
	"log"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

var db_mutex sync.Mutex

//NOTE: user_devices is name of table
func createUrlStatusTable(db *sql.DB) {
	createURLTableSQL := `CREATE TABLE user_devices (
		"username"			TEXT,
		"device_slice"		TEXT
	  );` // SQL Statement for Create Table

	fmt.Println("Create url_status table...")
	statement, err := db.Prepare(createURLTableSQL) // Prepare SQL Statement
	if err != nil {
		log.Fatal("Prepare crateURLTableSQL failed: " + err.Error())
	}
	defer statement.Close()
	statement.Exec() // Execute SQL Statements
	fmt.Println("url_status table created")
}

func DB_InsertDeviceSlice(db *sql.DB, username string, deviceslice []server.Device) {
	//fmt.Println("Inserting URL Status ...")
	insertUrlStatusSQL := `INSERT INTO user_devices(username, device_slice) VALUES (?,?)`
	db_mutex.Lock()
	statement, err := db.Prepare(insertUrlStatusSQL) // Prepare statement.
	// This is good to avoid SQL injections
	if err != nil {
		log.Fatalln("Prepare INSERT INTO failed: " + err.Error())
	}
	//timestamp := time.Now()
	//fmt.Printf("%s INSERT %s  %3d  %t   %s   %s   %s\n", digest_num, digest_num, line_in_digest, valid, timestamp, orig_url, derived_url)
	jsonstring, err := json.Marshal(&deviceslice)
	if err != nil {
		log.Printf("Marshal Error.\n")
		db_mutex.Unlock()
	}
	_, err = statement.Exec(username, jsonstring)
	db_mutex.Unlock()
	if err != nil {
		log.Printf("Execute INSERT INTO failed: %s\n", err.Error())
		time.Sleep(500 * time.Second)
	}
	statement.Close()
}

func DB_GetDeviceSlice(db *sql.DB, username string) ([]server.Device, error) {
	var err error
	var sqlrows *sql.Rows
	var jsondevslice string
	var retslice []server.Device
	//fmt.Println("Inserting URL Status ...")
	DB_selectdevslice := `SELECT device_slice from user_devices WHERE username == ?`
	db_mutex.Lock()
	sqlrows, err = db.Query(DB_selectdevslice, username) // Prepare statement.
	// This is good to avoid SQL injections
	if err != nil {
		log.Fatalln("SELECT failed: " + err.Error())
	}
	defer sqlrows.Close()
	db_mutex.Unlock()
	if sqlrows.Next() {
		sqlrows.Scan(&jsondevslice)
	}
	fmt.Printf("SELECTed entry: %s\n", jsondevslice)
	err = json.Unmarshal([]byte(jsondevslice), &retslice)
	if err != nil {
		log.Printf("failed unmarshal\n")
	}
	return retslice, err
}
