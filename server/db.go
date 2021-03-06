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

const currentDB string = "hermes"
const admin_profile = "prof"

var adminTable map[string]string = make(map[string]string)

var devdb_mutex sync.Mutex = sync.Mutex{}
var devicedb *sql.DB
var index_listing map[string]int

func DB_CreateAdminTable(db *sql.DB) (err error) {

	createAdminSQL := "USE " + currentDB
	create_statement, err := db.Prepare(createAdminSQL) // Prepare SQL Statement
	if err != nil {
		return err
	}
	defer create_statement.Close()
	create_statement.Exec()
	print(createAdminSQL)
	/*
		Admin profile should contain Admin Key, MySQL root cridentials
	*/
	createAdminSQL = `CREATE TABLE IF NOT EXISTS admin_profile (
			index   INT
			prof	TEXT
		)`

	create_statement, err = db.Prepare(createAdminSQL)
	Errhandle_Log(err, "create (db)")
	if err != nil {
		return err
	}
	defer create_statement.Close()
	create_statement.Exec()

	return

}

func UpdateAdminTable(db *sql.DB) (err error) {
	var jsondata string
	var prof map[string]string
	var admincode uuid.UUID
	//ordering: big endian
	var admincode_hexet_0 int64
	var admincode_hexet_1 int64

	var SELECT_admin_sql string = `SELECT prof FROM admin_profile WHERE index = ?`
	_, err = db.Exec(SELECT_admin_sql, 0)
	Errhandle_Log(err, ERRMSG_DB_SELECT)
	if err != nil {
		return err
	}

	json.Unmarshal([]byte(jsondata), &prof)
	//covers nil case
	if len(prof) == 0 {
		goto TABLE_DNE
	}
	err = json.Unmarshal([]byte(prof[admin_profile]), &admincode)
	Errhandle_Log(err, ERRMSG_JSON_UNMARSHALL)
	if err != nil {
		return err
	}
	admincode_hexet_0, admincode_hexet_1 = admincode.Time().UnixTime()
	if admincode_hexet_0 == 0 && admincode_hexet_1 == 0 {
		goto create_new_admincode
	}
	return nil

	//optional branches here. Each must terminate or goto another branch: The idea behind eulerian tours
TABLE_DNE:
	prof = make(map[string]string)
	goto create_new_admincode

create_new_admincode:
	admincode, err = uuid.NewUUID()
	Errhandle_Log(err, ERRMSG_CREATE_UUID)
	if err != nil {
		return err
	}
	prof[admin_profile] = admincode.String()
	goto MUTEX_ADMINTABLE_UPDATE

MUTEX_ADMINTABLE_UPDATE:
	var jsonstring []byte
	jsonstring, err = json.Marshal(prof)
	Errhandle_Log(err, ERRMSG_JSON_MARSHALL)
	if err != nil {
		return err
	}
	var UPDATE_STATEMENT string = `UPDATE admin_profile SET` + admin_profile + `= ?`
	UPDATE_CLOSER, err := db.Prepare(UPDATE_STATEMENT)
	Errhandle_Log(err, ERRMSG_DB_PREPARE_STATEMENT)
	if err != nil {
		return err
	}
	defer UPDATE_CLOSER.Close()
	devdb_mutex.Lock()
	_, err = UPDATE_CLOSER.Exec(string(jsonstring))
	Errhandle_Log(err, "DB Execute")
	devdb_mutex.Unlock()
	if err != nil {
		return err
	}
	return
}

//NOTE: user_db is name of table
func DB_CreateDeviceTable(db *sql.DB) error {

	createDatabaseSQL := "CREATE DATABASE IF NOT EXISTS " + currentDB
	create_statement, err := db.Prepare(createDatabaseSQL) // Prepare SQL Statement
	if err != nil {
		return err
	}
	defer create_statement.Close()
	create_statement.Exec() // Execute SQL Statements

	createURLTableSQL := `CREATE TABLE IF NOT EXISTS user_devices (
		username		TEXT,
		uuid			TEXT,
		userinfo		TEXT
	  )` // SQL Statement for Create Table

	//fmt.Println("Create url_status table...")
	statement, err := db.Prepare(createURLTableSQL) // Prepare SQL Statement
	if err != nil {
		return err
	}
	defer statement.Close()
	statement.Exec() // Execute SQL Statements
	if index_listing == nil {
		index_listing = make(map[string]int)
	}
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
	insertUrlStatusSQL := `INSERT INTO user_devices(username, uuid, userinfo) VALUES (?,?,?)`
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
	_, err = statement.Exec(username, USS.Userid.String(), jsonstring)
	if err != nil {
		log.Printf("Execute INSERT INTO failed: %s\n", err.Error())
		return err
	}
	index_listing[username] = 1
	statement.Close()
	return nil
}

/*
	Workflow:
	1) This will add `deviceslice` to the user with username `username` in the database
*/
func DB_AddDevSlice(db *sql.DB, username string, deviceslice map[uuid.UUID]Device) error {
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
	for _, v := range deviceslice {
		if index_listing[v.Username] != 0 {
			user, err := DB_GetUser(db, v.Username)
			if err != nil {
				return err
			}
			for i2, v2 := range deviceslice {
				user.Devicelist[i2] = v2
			}
		}
	}
	if done {
		return nil
	}

	//TODO: Add/append all devices that are in deviceslice but not in user.Devicelist to user.Devicelist
	//fmt.Println("Inserting URL Status ...")
	insertUrlStatusSQL := `INSERT INTO user_devices(username, uuid, userinfo) VALUES (?,?,?)`
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
	_, err = statement.Exec(username, user.User_uuid.String(), jsonstring)
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
	DB_selectdevslice := `SELECT userinfo from user_devices WHERE username = ?`
	devdb_mutex.Lock()
	defer devdb_mutex.Unlock()
	sqlrows, err = db.Query(DB_selectdevslice, username) // Prepare statement.
	Info_Log(username)
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

func DB_GetMultipleUsers(db *sql.DB, usernames []string) (retslice map[string]User, err error) {
	retslice = make(map[string]User)
	var USS UserStorageStruct
	for i, v := range usernames {
		if func() bool {
			var err error
			var sqlrows *sql.Rows
			var json_userstoragestruct_slice string
			//fmt.Println("Inserting URL Status ...")
			DB_selectdevslice := `SELECT userinfo from user_devices WHERE username = ?`
			devdb_mutex.Lock()
			defer devdb_mutex.Unlock()
			sqlrows, err = db.Query(DB_selectdevslice, usernames[i]) // Prepare statement.
			Errhandle_Log(err, ERRMSG_DB_SELECT)
			// This is good to avoid SQL injections
			if err != nil {
				goto END_DYNAMIC
			}
			defer sqlrows.Close()
			if sqlrows.Next() {
				sqlrows.Scan(&json_userstoragestruct_slice)
			} else {
				err = errors.New("not found")
				goto END_DYNAMIC
			}
			//fmt.Printf("SELECTed entry: %s\n", jsondevslice)
			if len(json_userstoragestruct_slice) == 0 {
				//there's this obscure way that this might make insertion errors... saying that null slices are just "not there". Trusting Go here. []=nil.
				err = errors.New("internal user retrieval error. Found entry but json slice was null")
				goto END_DYNAMIC
			}
			err = json.Unmarshal([]byte(json_userstoragestruct_slice), &USS)
			Errhandle_Log(err, ERRMSG_JSON_UNMARSHALL)
			if err != nil {
				log.Printf("failed unmarshal\n")
			}
		END_DYNAMIC:
			if err != nil {
				return false
			} else {
				return true
			}

		}() {
			retslice[v] = USS.UnPack()
		}
	}
	return
}

func DB_GetUserByUUID(db *sql.DB, uuid uuid.UUID) (User, error) {
	var err error
	var sqlrows *sql.Rows
	var json_userstoragestruct_slice string
	var USS UserStorageStruct
	var retUser User
	//fmt.Println("Inserting URL Status ...")
	DB_selectdevslice := `SELECT userinfo from user_devices WHERE uuid = ?`
	devdb_mutex.Lock()
	defer devdb_mutex.Unlock()
	sqlrows, err = db.Query(DB_selectdevslice, uuid.String()) // Prepare statement.
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
	if err != nil {
		return retSlice, err
	}
	Errhandle_Log(err, ERRMSG_DB_SELECT)
	Info_Log(baseUser)
	userMap = baseUser.Devicelist
	Info_Log(userMap)
	retSlice = make([]Device, len(userMap))
	for _, v := range userMap {
		retSlice[retSlice_index] = v
		retSlice_index++
	}
	return retSlice, nil
}

func DB_GetDeviceSliceByUUID(db *sql.DB, uid uuid.UUID) ([]Device, error) {
	var baseUser User
	var userMap map[uuid.UUID]Device
	var retSlice []Device
	var retSlice_index int = 0
	var err error
	baseUser, err = DB_GetUserByUUID(db, uid)
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
