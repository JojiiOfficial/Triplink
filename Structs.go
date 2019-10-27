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

//FetchRequest request strct for fetching changed ips
type FetchRequest struct {
	Token  string      `json:"token"`
	Filter FetchFilter `json:"filter"`
}

//FetchFilter to filter result from fetch request
type FetchFilter struct {
	Since        int64   `json:"since"`
	MinReason    float64 `json:"minReason"`
	MinReports   int     `json:"minReports"`
	ProxyAllowed int     `json:"allowProxy"`
	MaxIPs       uint    `json:"maxIPs"`
}

//FetchResponse struct for fetch response
type FetchResponse struct {
	IPs              []IPList `json:"ips"`
	CurrentTimestamp int64    `json:"cts"`
}

//IPList a list of ips from DB
type IPList struct {
	IP      string `json:"ip"`
	Deleted int    `json:"del"`
}
