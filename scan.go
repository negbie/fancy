package main

import (
	"bytes"
	"fmt"
	"time"
)

const seperator = ' '

var (
	errTemplate = fmt.Errorf("Unexpected rsyslog template format")
	errLevel    = fmt.Errorf("Unexpected rsyslog level format")
	errLength   = fmt.Errorf("Unexpected rsyslog message length")
)

func scanLine(raw []byte, metricOnly bool) (*LogLine, error) {
	var err error
	ll := &LogLine{
		Raw:       raw,
		Timestamp: time.Now(),
	}

	if len(ll.Raw) < 16 {
		return nil, errLength
	}

	if ll.Severity, err = getSeverity(raw[0]); err != nil {
		return nil, err
	}

	var curPos, endPos = 2, 2
	endPos = bytes.IndexRune(ll.Raw[curPos:], seperator)
	if endPos == -1 {

	}
	endPos += curPos
	ll.Hostname = string(ll.Raw[curPos:endPos])
	curPos = endPos + 1

	endPos = bytes.IndexRune(ll.Raw[curPos:], seperator)
	if endPos == -1 {
		return nil, errTemplate
	}
	endPos += curPos
	ll.Program = string(ll.Raw[curPos:endPos])
	curPos = endPos + 1

	if !ll.Valid() {
		return nil, errTemplate
	}

	if metricOnly {
		return ll, nil
	}

	ll.Msg = string(ll.Raw[curPos:])

	return ll, nil
}

func getSeverity(in byte) (out string, err error) {
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
		return "", errLevel
	}
	return out, nil
}
