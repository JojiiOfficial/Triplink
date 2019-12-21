package main

import (
	"net"
	"os"
	"os/exec"
	"time"
)

var reservedIPs = []string{
	"0.0.0.0/8",
	"10.0.0.0/8",
	"127.0.0.0/8",
	"169.254.0.0/16",
	"172.16.0.0/12",
	"192.0.0.0/24",
	"192.0.2.0/24",
	"192.88.99.0/24",
	"192.168.0.0/16",
	"198.18.0.0/15",
	"224.0.0.0/4",
	"240.0.0.0/4",
}

//returns if ip is valid and a reason
func isIPValid(ip string) (bool, int) {
	pip := net.ParseIP(ip)
	if pip.To4() == nil {
		return false, 0
	}
	for _, reservedIP := range reservedIPs {
		_, subnet, err := net.ParseCIDR(reservedIP)
		if err != nil {
			panic(err)
		}
		if subnet.Contains(pip) {
			return false, -1
		}
	}
	return true, 1
}

func ipErrToString(err int) string {
	switch err {
	case 1:
		{
			return "Succes"
		}
	case -1:
		{
			return "Reserved"
		}
	case 0:
		{
			return "No ipv4"
		}
	}
	return ""
}

func appendLogs(newf, logs string) {
	file, err := os.OpenFile(newf, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 755)
	if err != nil {
		panic(err)
	}
	_, err = file.WriteString(logs + "\n")
	if err != nil {
		panic(err)
	}
	file.Close()
}

func runCommand(errorHandler func(error, string), sCmd string) (outb string, err error) {
	out, err := exec.Command("sh", "-c", sCmd).Output()
	output := string(out)
	if err != nil {
		if errorHandler != nil {
			errorHandler(err, sCmd)
		}
		return "", err
	}
	return output, nil
}

func cidrToIPlist(cidr string) ([]string, error) {
	ip, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, err
	}

	var ips []string
	for ip := ip.Mask(ipnet.Mask); ipnet.Contains(ip); incIP(ip) {
		ips = append(ips, ip.String())
	}
	return ips, nil
}

func incIP(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}

func parseTimeStamp(unix int64) string {
	return time.Unix(unix, 0).Format(time.Stamp)
}
