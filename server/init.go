package main

import "database/sql"

func InitiateEverything() *sql.DB {
	var db *sql.DB = InitDeviceTables()
	return db
	//Deprecated
	//hellabackend.InitPortGenerator(PORT_LOWER_BOUND, PORT_UPPER_BOUND)

}
