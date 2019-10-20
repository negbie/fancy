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
	parsed := []byte(fmt.Sprintf("%s %s %s %s %s",
		setSeverity(l.Severity), l.Hostname, l.Facility, l.Program, l.Msg))
	return bytes.Equal(l.Raw, parsed)
}

func setSeverity(in string) (out string) {
	switch in {
	case "emergency":
		out = "0"
	case "alert":
		out = "1"
	case "critical":
		out = "2"
	case "error":
		out = "3"
	case "warning":
		out = "4"
	case "notice":
		out = "5"
	case "info":
		out = "6"
	case "debug":
		out = "7"
	default:
		out = "0"
	}
	return
}
