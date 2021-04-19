package main

import (
	"database/sql"
	"fmt"
	_ "hermes/server"
	"log"
	"sync"
	"time"
)

var db_mutex sync.Mutex

//NOTE: user_devices is name of table
func createUrlStatusTable(db *sql.DB) {
	createURLTableSQL := `CREATE TABLE user_devices (
		"DeviceSlice"		TEXT,
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

func insertUrlStatus(db *sql.DB, deviceslice []device) {
	//fmt.Println("Inserting URL Status ...")
	insertUrlStatusSQL := `INSERT INTO user_devices(DeviceSlice) VALUES (?)`
	db_mutex.Lock()
	statement, err := db.Prepare(insertUrlStatusSQL) // Prepare statement.
	// This is good to avoid SQL injections
	if err != nil {
		log.Fatalln("Prepare INSERT INTO failed: " + err.Error())
	}
	timestamp := time.Now()
	fmt.Printf("%s INSERT %s  %3d  %t   %s   %s   %s\n", digest_num, digest_num, line_in_digest, valid, timestamp, orig_url, derived_url)
	_, err = statement.Exec(digest_num, line_in_digest, valid, timestamp, orig_url, derived_url)
	db_mutex.Unlock()
	if err != nil {
		log.Printf("Execute INSERT INTO failed: " + err.Error())
		time.Sleep(500 * time.Second)
	}
	statement.Close()
}
