package main

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type Token struct {
	UserID        string
	Val           []byte
	TimeGenerated time.Time
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
func (u *User) FetchUserLinkRequestByUserID() {}

//enable a user to request permission from another user to ask to send them
//content
//NOTICE: DENIED. MUST ACCEPT.
func (u *User) Preapprove_link(username string) error {
	return denied
}
func (u *User) Rescind_preapproval(username string) error {
	return denied
}
