package main

import "github.com/google/uuid"

type user struct {
	Username   string
	Userid     uuid.UUID
	Devicelist map[uuid.UUID]device
}

func NewUser(username string) (retuser user, err error) {
	id, err := uuid.NewUUID()
	retuser = user{Username: username, Userid: id, Devicelist: make(map[uuid.UUID]device)}
	return
}
