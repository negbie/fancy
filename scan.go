package main

import (
	"bytes"
	"fmt"
	"time"
)

var errScan = fmt.Errorf("Unexpected Log Line Format")

func scanLine(raw []byte) (*LogLine, error) {
	var curPos, endPos = 2, 2
	lm := &LogLine{Raw: raw}
	lm.Timestamp = time.Now()
	lm.Severity = severityText(lm.Raw[0])

	endPos = bytes.IndexRune(lm.Raw[curPos:], ' ')
	if endPos == -1 {
		return nil, errScan
	}
	endPos += curPos
	lm.Hostname = string(lm.Raw[curPos:endPos])
	curPos = endPos + 1

	endPos = bytes.IndexRune(lm.Raw[curPos:], ' ')
	if endPos == -1 {
		return nil, errScan
	}
	endPos += curPos
	lm.Facility = string(lm.Raw[curPos:endPos])
	curPos = endPos + 1

	endPos = bytes.IndexRune(lm.Raw[curPos:], ' ')
	if endPos == -1 {
		return nil, errScan
	}
	endPos += curPos
	lm.Program = string(lm.Raw[curPos:endPos])
	curPos = endPos + 1

	lm.Msg = string(lm.Raw[curPos:])

	return lm, nil
}

func severityText(severity byte) (text string) {
	switch int(severity) {
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
