// Command homedns is a utility to update a Linode managed DNS A record with
// the system's public IP address.
//
// Usage of homedns:
//   -domain string
//     	DNS Domain name, required
//   -key string
//     	Linode API key, required
//   -name string
//     	DNS A Record name, required
//   -verbose
//     	Enable verbose logging
//   -help
//     	Show this help text
package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/timewasted/linode"
	"github.com/timewasted/linode/dns"
)

var apiKey = flag.String("key", "", "Linode API key, required")
var dnsDomainName = flag.String("domain", "", "DNS Domain name, required")
var dnsARecordName = flag.String("name", "", "DNS A Record name, required")
var verbose = flag.Bool("verbose", false, "Enable verbose logging")

var ipv4Matcher = regexp.MustCompile(`([0-9]{0,3}\.[0-9]{0,3}\.[0-9]{0,3}\.[0-9]{0,3})`)

func UpdateDomainResourceTarget(ldns *dns.DNS, r *dns.Resource, target string) error {
	params := linode.Parameters{
		"DomainID":   strconv.Itoa(r.DomainID),
		"ResourceID": strconv.Itoa(r.ResourceID),
		"Target":     target,
	}
	lin := ldns.ToLinode()
	_, err := lin.Request("domain.resource.update", params, nil)

	return err
}

func IsPublicIP(ip net.IP) bool {
	if ip.IsLoopback() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() {
		return false
	}
	p := ip.To4()
	if p == nil {
		return false
	}
	switch true {
	case p[0] == 10:
		return false
	case p[0] == 172 && p[1] >= 16 && p[1] <= 31:
		return false
	case p[0] == 192 && p[1] == 168:
		return false
	}
	return true
}

func GetPublicIP() (net.IP, error) {
	resp, err := http.Get("http://checkip.dyndns.org")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	m := ipv4Matcher.FindAll(body, -1)
	if len(m) != 1 {
		return nil, fmt.Errorf("couldnt parse response: %s", string(body))
	}
	ip := string(m[0])
	parsed := net.ParseIP(ip)
	if parsed == nil {
		return nil, fmt.Errorf("couldnt parse IP %s", ip)
	}
	if !IsPublicIP(parsed) {
		return nil, fmt.Errorf("%s is a private IP", ip)
	}
	return parsed, nil
}

func init() {
	flag.Usage = func() {
		fmt.Println(`homedns is a utility to update a Linode managed DNS A record with the system's 
public IP address.

Usage of homedns:`)
		flag.PrintDefaults()
		fmt.Println(`  -help
    	Show this help text`)
	}
	flag.Parse()
	var missing []string
	if *apiKey == "" {
		missing = append(missing, "key")
	}
	if *dnsDomainName == "" {
		missing = append(missing, "domain")
	}
	if *dnsARecordName == "" {
		missing = append(missing, "name")
	}
	if len(missing) > 0 {
		fatalf("Error: missing required parameters: %s\n", strings.Join(missing, ", "))
	}
}

func main() {
	publicIP, err := GetPublicIP()
	if err != nil {
		fatalf("Error getting public IP: %s\n", err.Error())
	}
	debugf("Public IP: %s\n", publicIP.String())

	l := dns.New(*apiKey)
	domain, err := l.GetDomain(*dnsDomainName)
	if err != nil {
		fatalf("Error getting DNS settings: %s", err.Error())
	}
	debugf("Pointing %s.%s to %s\n", *dnsARecordName, domain.Domain, publicIP.String())

	recs, err := l.GetResourcesByType(domain.DomainID, "A")
	if err != nil {
		fatalf("Error getting domain records: %s", err.Error())
	}

	var existing *dns.Resource
	for _, rec := range recs {
		if rec.Name == *dnsARecordName {
			existing = rec
			break
		}
	}
	if existing == nil {
		_, err = l.CreateDomainResourceA(domain.DomainID, *dnsARecordName, publicIP.String(), 300)
		if err != nil {
			fatalf("Error creating DNS record: %s", err.Error())
		}
		debugln("Successfully created DNS record.")

	} else if publicIP.String() != existing.Target {
		err := UpdateDomainResourceTarget(l, existing, publicIP.String())
		if err != nil {
			fatalf("Error updating DNS record: %s", err.Error())
		}
		debugln("Successfully updated DNS record.")

	} else {
		debugln("Existing DNS record is correct, exiting successfully.")
	}
}
