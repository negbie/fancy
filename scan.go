package main

import (
	"bytes"
	"fmt"
	"time"
)

const seperator = ' '

var errScan = fmt.Errorf("Unexpected rsyslog template format")

func scanLine(raw []byte, metricOnly bool) (*LogLine, error) {
	ll := &LogLine{
		Raw:       raw,
		Timestamp: time.Now(),
		Severity:  getSeverity(raw[0]),
	}

	var curPos, endPos = 2, 2
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
	ll.Program = string(ll.Raw[curPos:endPos])
	curPos = endPos + 1

	if metricOnly {
		return ll, nil
	}

	ll.Msg = string(ll.Raw[curPos:])

	return ll, nil
}

func getSeverity(in byte) (out string) {
	switch in {
	case 48: // 0
		out = "emergency"
	case 49: // 1
		out = "alert"
	case 50: // 2
		out = "critical"
	case 51: // 3
		out = "error"
	case 52: // 4
		out = "warning"
	case 53: // 5
		out = "notice"
	case 54: // 6
		out = "info"
	case 55: // 7
		out = "debug"
	default:
		out = "unknown"
	}
	return
}
