package main

import (
	"bufio"
	"fmt"
	"log"
	"log/syslog"
	"os"
)

func main() {
	s := bufio.NewScanner(os.Stdin)
	for s.Scan() {
		lm, err := scanLine(s.Bytes())
		if err != nil {
			return
		}
		fmt.Fprintf(os.Stderr, lm.String())

	}
}

func ping() {
	lw, err := syslog.Dial("tcp", "localhost:514", syslog.LOG_DEBUG, "fancy")
	if err != nil {
		log.Fatal(err)
	}
	lw.Info("ping from fancy!")
}
