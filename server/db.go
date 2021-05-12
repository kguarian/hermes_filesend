package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"sync"

	"github.com/google/uuid"
)

var devdb_mutex sync.Mutex = sync.Mutex{}
var devicedb *sql.DB

//NOTE: user_db is name of table
func DB_CreateDeviceTable(db *sql.DB) error {
	createURLTableSQL := `CREATE TABLE user_devices (
		"username"			TEXT,
		"userinfo"		TEXT
	  );` // SQL Statement for Create Table

	//fmt.Println("Create url_status table...")
	statement, err := db.Prepare(createURLTableSQL) // Prepare SQL Statement
	if err != nil {
		return err
	}
	defer statement.Close()
	statement.Exec() // Execute SQL Statements
	fmt.Println("user_devices table created")
	return nil
}

func DB_AddUser(db *sql.DB, username string, user User) error {
	var userToAdd User
	var USS UserStorageStruct
	var err error
	userToAdd, err = DB_GetUser(db, username)
	Errhandle_Log(err, err.Error())
	if err == nil {
		return errors.New(ERRMSG_DB_ATTEMPTED_INSERT_DUPLICATE)
	}
	userToAdd = user
	USS = userToAdd.Store()
	//fmt.Println("Inserting URL Status ...")
	insertUrlStatusSQL := `INSERT INTO user_devices(username, userinfo) VALUES (?,?)`
	//timestamp := time.Now()
	//fmt.Printf("%s INSERT %s  %3d  %t   %s   %s   %s\n", digest_num, digest_num, line_in_digest, valid, timestamp, orig_url, derived_url)
	jsonstring, err := json.Marshal(&USS)
	if err != nil {
		log.Printf("Marshal Error.\n")
	}
	devdb_mutex.Lock()
	defer devdb_mutex.Unlock()
	statement, err := db.Prepare(insertUrlStatusSQL) // Prepare statement.
	// This is good to avoid SQL injections
	if err != nil {
		return err
	}
	_, err = statement.Exec(username, jsonstring)
	if err != nil {
		log.Printf("Execute INSERT INTO failed: %s\n", err.Error())
		return err
	}
	statement.Close()
	return nil
}

/*
	Workflow:
	1) Get User struct by `username`
		1a) If DNE, then create it. Now we have User struct
	2) check for device in user struct device slice.
		2a) If exists, then we're done.
		2b) If DNE, then we create it, then we add it to devslice,
			then we insert into table, then we're done.
*/
func DB_AddDevSlice(db *sql.DB, username string, deviceslice []Device) error {
	user, err := DB_GetUser(db, username)
	Errhandle_Log(err, ERRMSG_DB_SELECT)
	if err != nil {
		user, err = NewUser(username)
		Errhandle_Log(err, err.Error())
		if err != nil {
			return err
		}
	}
	//see if we need to even do anything, return if not
	done := true
	deviceinslice := make([]bool, len(deviceslice))
	for _, v1 := range user.Devicelist {
		for i, v2 := range deviceslice {
			if v1.Equal(&v2) {
				deviceinslice[i] = true
			}
		}
	}
	for i, v := range deviceinslice {
		Info_Log(v)
		if !v {
			user.Devicelist[deviceslice[i].Device_uuid] = deviceslice[i]

			Info_Log(user.Devicelist)
			done = false
		}
	}
	if done {
		return nil
	}

	//TODO: Add/append all devices that are in deviceslice but not in user.Devicelist to user.Devicelist
	//fmt.Println("Inserting URL Status ...")
	insertUrlStatusSQL := `INSERT INTO user_devices(username, userinfo) VALUES (?,?)`
	//timestamp := time.Now()
	//fmt.Printf("%s INSERT %s  %3d  %t   %s   %s   %s\n", digest_num, digest_num, line_in_digest, valid, timestamp, orig_url, derived_url)
	Info_Log(user)
	jsonstring, err := json.Marshal(&user)
	Errhandle_Log(err, ERRMSG_JSON_MARSHALL)
	if err != nil {
		log.Printf("Marshal Error.\n")
	}
	statement, err := db.Prepare(insertUrlStatusSQL) // Prepare statement.
	Errhandle_Log(err, ERRMSG_DB_PREPARE_STATEMENT)
	if err != nil {
		return err
	}

	devdb_mutex.Lock()
	defer devdb_mutex.Unlock()
	// This is good to avoid SQL injections
	_, err = statement.Exec(username, jsonstring)
	if err != nil {
		log.Printf("Execute INSERT INTO failed: %s\n", err.Error())
		return err
	}
	statement.Close()
	return nil
}

func DB_GetUser(db *sql.DB, username string) (User, error) {
	var err error
	var sqlrows *sql.Rows
	var json_userstoragestruct_slice string
	var USS UserStorageStruct
	var retUser User
	//fmt.Println("Inserting URL Status ...")
	DB_selectdevslice := `SELECT userinfo from user_devices WHERE username == ?`
	devdb_mutex.Lock()
	defer devdb_mutex.Unlock()
	sqlrows, err = db.Query(DB_selectdevslice, username) // Prepare statement.
	Errhandle_Log(err, ERRMSG_DB_SELECT)
	// This is good to avoid SQL injections
	if err != nil {
		return retUser, err
	}
	defer sqlrows.Close()
	if sqlrows.Next() {
		sqlrows.Scan(&json_userstoragestruct_slice)
	} else {
		return retUser, errors.New("not found")
	}
	//fmt.Printf("SELECTed entry: %s\n", jsondevslice)
	if len(json_userstoragestruct_slice) == 0 {
		return retUser, errors.New("internal user retrieval error. Found entry but json slice was null")
	}
	err = json.Unmarshal([]byte(json_userstoragestruct_slice), &USS)
	Errhandle_Log(err, ERRMSG_JSON_UNMARSHALL)
	if err != nil {
		log.Printf("failed unmarshal\n")
	}
	return USS.UnPack(), err
}

//err being nil does not imply that the returned device slice is not nil.
//err being non-nil implies that you should not use the returned slice.
func DB_GetDeviceSlice(db *sql.DB, username string) ([]Device, error) {
	var baseUser User
	var userMap map[uuid.UUID]Device
	var retSlice []Device
	var retSlice_index int = 0
	var err error
	baseUser, err = DB_GetUser(db, username)
	Errhandle_Log(err, ERRMSG_DB_SELECT)
	Info_Log(baseUser)
	if err != nil {
		return retSlice, err
	}
	userMap = baseUser.Devicelist
	Info_Log(userMap)
	retSlice = make([]Device, len(userMap))
	for _, v := range userMap {
		retSlice[retSlice_index] = v
		retSlice_index++
	}
	return retSlice, nil
}
