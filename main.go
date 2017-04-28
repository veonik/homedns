package main

import (
	"flag"
	"log"
	"net"
	"os/exec"
	"strings"

	"github.com/timewasted/linode/dns"
)

var apiKey = flag.String("key", "", "Linode API key.")
var dnsDomainName = flag.String("domain", "", "DNS Domain Name")
var dnsARecordName = flag.String("name", "", "DNS A Record Name")

func main() {
	flag.Parse()
	if *apiKey == "" {
		log.Fatalln("Missing required parameter: key")
	}
	if *dnsDomainName == "" {
		log.Fatalln("Missing required parameter: domain")
	}
	if *dnsARecordName == "" {
		log.Fatalln("Missing required parameter: name")
	}

	out, err := exec.Command("dig", "TXT", "+short", "o-o.myaddr.l.google.com", "@ns1.google.com").Output()
	if err != nil {
		log.Fatalf("Unable to get public IP: %s\n", err.Error())
	}
	ip := strings.Trim(string(out), "\"\n")
	log.Printf("Public IP: %s\n", ip)

	publicIP := net.ParseIP(ip)
	if publicIP == nil {
		log.Fatalf("Invalid Public IP: %s\n", ip)
	}

	l := dns.New(*apiKey)
	domain, err := l.GetDomain(*dnsDomainName)
	if err != nil {
		log.Fatalf("Error getting DNS settings: %s", err.Error())
	}
	log.Printf("Pointing %s.%s to %s\n", *dnsARecordName, domain.Domain, publicIP.String())

	_, err = l.CreateDomainResourceA(domain.DomainID, *dnsARecordName, publicIP.String(), 300)
	if err != nil {
		log.Fatalf("Error updating DNS record: %s", err.Error())
	}
	log.Println("Successfully updated DNS record.")
}
