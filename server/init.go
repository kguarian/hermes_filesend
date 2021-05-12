package main

import "database/sql"

func InitiateEverything() (*sql.DB, chan error) {
	var reterrs chan error = make(chan error)
	var err error
	var db *sql.DB = InitDeviceTables()
	err = selectServerIP()
	if err != nil {
		reterrs <- err
	}
	return db, reterrs
	//Deprecated
	//hellabackend.InitPortGenerator(PORT_LOWER_BOUND, PORT_UPPER_BOUND)

}
