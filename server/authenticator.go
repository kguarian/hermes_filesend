package main

//enable a user to request permission from another user to ask to send them
//content
func (u *User) RequestUserLink(tgtUID string) {

}

//Returns data structure containing all User LInk Requests
func (u *User) FetchUserLinkRequests() {}

//Returns single UserLink if it's matched, null UserInfo if else
func (u *User) FetchUserLinkRequestByUserID() {}
