package main

//IPset a report set containing ip and a reason
type IPset struct {
	IP     string `json:"ip"`
	Reason int    `json:"r"`
}

//ReportIPStruct incomming ip report
type ReportIPStruct struct {
	Token string  `json:"token"`
	Ips   []IPset `json:"ips"`
}
