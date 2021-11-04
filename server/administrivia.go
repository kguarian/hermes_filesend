package main

import "github.com/google/uuid"

var adminkey uuid.UUID

func GenerateAdminKey() {
	var err error

	adminkey, err = uuid.NewUUID()

	Errhandle_Log(err, "FAILED TO GENERATE ADMIN KEY")
}
