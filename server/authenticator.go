package main

import (
	"errors"

	"github.com/google/uuid"
)

//enable a user to request permission from another user to ask to send them
//content
func (u *User) Preapprove_link(username string) error {
	user, err := DB_GetUser(devicedb, username)
	Errhandle_Log(err, ERRMSG_DB_SELECT)
	if err != nil {
		return err
	}
	u.Preapproved_users <- user.User_uuid
	return nil
}

func (u *User) Rescind_preapproval(username string) error {
	var device uuid.UUID
	var preaplist chan uuid.UUID = u.Preapproved_users
	var tgtuuid uuid.UUID

	preappsz := len(u.Preapproved_users)
	if preappsz == 0 {
		return errors.New("no preapproval to rescind")
	}

	user, err := DB_GetUser(devicedb, username)
	Errhandle_Log(err, ERRMSG_DB_SELECT)
	if err != nil {
		return err
	}
	tgtuuid = user.User_uuid

	for i := 0; i < preappsz; i++ {
		device = <-preaplist
		if device == tgtuuid {
			return nil
		}
		preaplist <- device
	}
	return errors.New(ERRMSG_DEVICENOTFOUND)
}

func (u *User) SendLinkRequest(username string) error {
	var device uuid.UUID
	var preaplist chan uuid.UUID
	var tgtuuid uuid.UUID
	var preappsz int

	user, err := DB_GetUser(devicedb, username)
	Errhandle_Log(err, ERRMSG_DB_SELECT)
	if err != nil {
		return err
	}
	preaplist = user.Preapproved_users
	preappsz = len(preaplist)
	if preappsz == 0 {
		return errors.New("no preapprovals; cannot Send Link Request")
	}

	tgtuuid = u.User_uuid

	for i := 0; i < preappsz; i++ {
		device = <-preaplist
		if device == tgtuuid {
			user.Linkrequests <- tgtuuid
			preaplist <- device
		}
		preaplist <- device
	}
	return errors.New(ERRMSG_DEVICENOTFOUND)
}

//Returns data structure containing all User Link Requests.
func (u *User) FetchUserLinkRequests() []User {
	var lenlinkreq int = len(u.Linkrequests)
	var retSlice []User = make([]User, lenlinkreq)
	var sliceindex int = 0
	var curruuid uuid.UUID
	for i := 0; i < lenlinkreq; i++ {
		curruuid = <-u.Linkrequests
		user, err := DB_GetUserByUUID(devicedb, curruuid)
		if err != nil {
			continue
		}

		retSlice[sliceindex] = user
		sliceindex++

		u.Linkrequests <- curruuid
	}
	return retSlice
}

//why?

// //Returns single UserLink if it's matched, null UserInfo if else
// func (u *User) FetchUserLinkRequestByUserID() {}
