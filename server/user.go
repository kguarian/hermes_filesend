package main

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
)

type User struct {
	Username    string
	User_uuid   uuid.UUID
	Devicelist  map[uuid.UUID]Device
	Inrequests  chan contentinfo       `json:"-"`
	Consentlist map[string]contentinfo `json:"clist"`
}

type UserStorageStruct struct {
	Username    string
	Userid      uuid.UUID
	Devicelist  map[uuid.UUID]Device
	Inrequests  []contentinfo
	Consentlist map[string]contentinfo `json:"clist"`
}

func (u *User) Store() UserStorageStruct {
	slicelen := len(u.Inrequests)
	var exchslice []contentinfo = make([]contentinfo, slicelen)
	for i := 0; i < slicelen; i++ {
		exchslice[i] = <-u.Inrequests
	}
	var retStruct UserStorageStruct = UserStorageStruct{Username: u.Username, Userid: u.User_uuid, Devicelist: u.Devicelist, Inrequests: exchslice, Consentlist: u.Consentlist}
	return retStruct
}

func (uss *UserStorageStruct) UnPack() User {
	var exchchan chan contentinfo = make(chan contentinfo)
	for i, _ := range uss.Inrequests {
		exchchan <- uss.Inrequests[i]
	}
	var retStruct User = User{Username: uss.Username, User_uuid: uss.Userid, Devicelist: uss.Devicelist, Inrequests: exchchan, Consentlist: uss.Consentlist}
	return retStruct

}

func NewUser(username string) (retuser User, err error) {
	id, err := uuid.NewUUID()
	retuser = User{Username: username, User_uuid: id, Devicelist: make(map[uuid.UUID]Device)}
	return
}

//TODO: Networked form
func (u *User) EvalConsent(sender *Device) error {
	var character byte
	var cont contentinfo
	var err error
	var ok bool

	for len(u.Inrequests) > 0 {
		cont, ok = <-u.Inrequests
		fmt.Printf("%s channel depth: %d\n", u.MarshalUser(), len(u.Inrequests))
		if !ok {
			err = errors.New(ERRMSG_CHANNEL_OPERATION)
			return err
		}
		if cont.Senderid == sender.Device_uuid {
			fmt.Printf("%s requests to send file [%s], size: %d bytes.\n", cont.Senderid, cont.Name, cont.Sizebytes)
			fmt.Printf("Approve? (Y/*): ")
			fmt.Scanf("%c\n", &character)
			fmt.Println(character == 'Y')
			u.Consentlist[cont.Name] = cont
		}
	}
	return nil
}

//This is an incredibly useless wrapper function that harms no one.
func (u *User) MarshalUser() []byte {
	retinfo, err := json.Marshal(u)
	Errhandle_Log(err, ERRMSG_JSON_MARSHALL)
	if err != nil {
		return nil
	}
	return retinfo
}
