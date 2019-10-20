package main

import (
	"bytes"
	"fmt"
	"time"
)

const seperator = ' '

var errScan = fmt.Errorf("Unexpected rsyslog template format")

func scanLine(raw []byte) (*LogLine, error) {
	var curPos, endPos = 2, 2
	ll := &LogLine{Raw: raw}
	ll.Timestamp = time.Now()
	ll.Severity = severityText(ll.Raw[0])

	endPos = bytes.IndexRune(ll.Raw[curPos:], seperator)
	if endPos == -1 {
		return nil, errScan
	}
	endPos += curPos
	ll.Hostname = string(ll.Raw[curPos:endPos])
	curPos = endPos + 1

	endPos = bytes.IndexRune(ll.Raw[curPos:], seperator)
	if endPos == -1 {
		return nil, errScan
	}
	endPos += curPos
	ll.Facility = string(ll.Raw[curPos:endPos])
	curPos = endPos + 1

	endPos = bytes.IndexRune(ll.Raw[curPos:], seperator)
	if endPos == -1 {
		return nil, errScan
	}
	endPos += curPos
	ll.Program = string(ll.Raw[curPos:endPos])
	curPos = endPos + 1

	ll.Msg = string(ll.Raw[curPos:])

	return ll, nil
}

func severityText(severity byte) (text string) {
	switch severity {
	case 48: // 0
		text = "emergency"
	case 49: // 1
		text = "alert"
	case 50: // 2
		text = "critical"
	case 51: // 3
		text = "error"
	case 52: // 4
		text = "warning"
	case 53: // 5
		text = "notice"
	case 54: // 6
		text = "info"
	case 55: // 7
		text = "debug"
	default:
		text = "unknown"
	}
	return
}
