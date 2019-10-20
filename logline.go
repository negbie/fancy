package main

import (
	"bytes"
	"fmt"
	"time"
)

type LogLine struct {
	Timestamp time.Time
	Severity  string
	Hostname  string
	Facility  string
	Program   string
	Msg       string
	Raw       []byte
}

func (l LogLine) String() string {
	return fmt.Sprintf("time:%v severity:%s hostname:%s facility:%s program:%s msg:%s raw:%s\n",
		l.Timestamp, l.Severity, l.Hostname, l.Facility, l.Program, l.Msg, l.Raw)
}

func (l LogLine) Valid() bool {
	parsed := []byte(fmt.Sprintf("%s %s %s %s %s", severityByte(l.Severity), l.Hostname, l.Facility, l.Program, l.Msg))
	return bytes.Equal(l.Raw, parsed)
}

func severityByte(a string) (b string) {
	switch a {
	case "emergency":
		b = "0"
	case "alert":
		b = "1"
	case "critical":
		b = "2"
	case "error":
		b = "3"
	case "warning":
		b = "4"
	case "notice":
		b = "5"
	case "info":
		b = "6"
	case "debug":
		b = "7"
	default:
		b = "0"
	}
	return
}
