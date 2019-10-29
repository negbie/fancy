package main

import (
	"bytes"
	"fmt"
	"time"
)

const seperator = ' '

var (
	errTemplate = fmt.Errorf("Unexpected rsyslog template format")
	errTime     = fmt.Errorf("Unexpected rsyslog time format")
	errLevel    = fmt.Errorf("Unexpected rsyslog level format")
	errLength   = fmt.Errorf("Unexpected rsyslog message length")
)

func parseLine(raw []byte, promOnly bool) (*LogLine, error) {
	var err error
	ll := &LogLine{
		Raw: raw,
	}

	if len(ll.Raw) < 46 {
		return nil, errLength
	}

	if !promOnly {
		ll.Timestamp, err = time.Parse((time.RFC3339), string(ll.Raw[:32]))
		if err != nil {
			return nil, errTime
		}
	}

	if ll.Severity, err = getSeverity(ll.Raw[33]); err != nil {
		return nil, err
	}

	var curPos, endPos = 35, 35
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
	ll.MsgPos = curPos

	if !ll.Valid() {
		return nil, errTemplate
	}

	if !promOnly {
		ll.Msg = string(ll.Raw[ll.MsgPos:])
	}

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
