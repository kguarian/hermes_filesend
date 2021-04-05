package main

import "github.com/google/uuid"

type user struct {
	devicelist map[uuid.UUID]device
}
