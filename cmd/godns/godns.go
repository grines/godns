package godns

import (
	"log"
	"strconv"

	completion "github.com/grines/godns/internal/app/completions"
	"github.com/miekg/dns"
)

func Start() {
	go completion.Start()

	dns.HandleFunc("dns.gostripe.click.", completion.HandleDnsRequest)

	// start server
	port := 8888
	server := &dns.Server{Addr: ":" + strconv.Itoa(port), Net: "udp"}
	log.Printf("Starting at %d\n", port)
	err := server.ListenAndServe()
	defer server.Shutdown()
	if err != nil {
		log.Fatalf("Failed to start server: %s\n ", err.Error())
	}
}
