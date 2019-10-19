package main

import (
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
	return fmt.Sprintf("time:%v\nseverity:%s\nhostname:%s\nfacility:%s\nprogram:%s\nmsg:%s\nraw:%s\n", l.Timestamp, l.Severity, l.Hostname, l.Facility, l.Program, l.Msg, l.Raw)
}
