package main

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
)

type User struct {
	Username          string
	User_uuid         uuid.UUID
	Devicelist        map[uuid.UUID]Device
	Preapproved_users chan uuid.UUID
	Linkrequests      chan uuid.UUID
	Inrequests        chan contentinfo       `json:"-"`
	Consentlist       map[string]contentinfo `json:"clist"`
}

type UserStorageStruct struct {
	Username          string
	Userid            uuid.UUID
	Devicelist        map[uuid.UUID]Device
	Preapproved_users []uuid.UUID
	Linkrequests      []uuid.UUID
	Inrequests        []contentinfo
	Consentlist       map[string]contentinfo `json:"clist"`
}

func (u *User) Store() UserStorageStruct {
	inreqlen := len(u.Inrequests)
	preapplen := len(u.Preapproved_users)
	linkreqlen := len(u.Linkrequests)
	var exchslice []contentinfo = make([]contentinfo, inreqlen)
	var expreapp []uuid.UUID = make([]uuid.UUID, preapplen)
	var exlinkreq []uuid.UUID = make([]uuid.UUID, linkreqlen)
	for i := 0; i < inreqlen; i++ {
		exchslice[i] = <-u.Inrequests
	}
	for i := 0; i < preapplen; i++ {
		expreapp[i] = <-u.Preapproved_users
	}
	for i := 0; i < linkreqlen; i++ {
		exlinkreq[i] = <-u.Linkrequests
	}
	var retStruct UserStorageStruct = UserStorageStruct{Username: u.Username, Userid: u.User_uuid, Devicelist: u.Devicelist, Preapproved_users: expreapp, Linkrequests: exlinkreq, Inrequests: exchslice, Consentlist: u.Consentlist}
	return retStruct
}

func (uss *UserStorageStruct) UnPack() User {
	var ex_inreq chan contentinfo = make(chan contentinfo)
	var ex_preapp chan uuid.UUID = make(chan uuid.UUID)
	var ex_linkreq chan uuid.UUID = make(chan uuid.UUID)
	for i, _ := range uss.Inrequests {
		ex_inreq <- uss.Inrequests[i]
	}
	for i, _ := range uss.Preapproved_users {
		ex_preapp <- uss.Preapproved_users[i]
	}
	for i, _ := range uss.Preapproved_users {
		ex_linkreq <- uss.Linkrequests[i]
	}
	var retStruct User = User{Username: uss.Username, User_uuid: uss.Userid, Devicelist: uss.Devicelist, Preapproved_users: ex_preapp, Linkrequests: ex_linkreq, Inrequests: ex_inreq, Consentlist: uss.Consentlist}
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
