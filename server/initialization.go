package main

import "database/sql"

func InitiateEverything(error_channel chan error) (retDB *sql.DB) {
	var err error
	retDB = InitDeviceTables()
	err = selectServerIP()
	if err != nil {
		error_channel <- err
	}
	//Deprecated
	//hellabackend.InitPortGenerator(PORT_LOWER_BOUND, PORT_UPPER_BOUND)
	return
}
