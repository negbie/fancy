package main

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_scanLine(t *testing.T) {
	raw := []byte("6 pad kern fancy {\"key1\":\"val1\", \"key2\":\"val2\"}")
	severity := "info"
	hostname := "pad"
	facility := "kern"
	program := "fancy"
	msg := "{\"key1\":\"val1\", \"key2\":\"val2\"}"
	res := bytes.NewBufferString("")

	scanLine(raw)
	assert.Equal(t, 0, res.Len())

	res.Reset()
	lm, err := scanLine(raw)
	if err != nil {
		t.Fail()
	}
	assert.Equal(t, severity, lm.Severity)
	assert.Equal(t, hostname, lm.Hostname)
	assert.Equal(t, facility, lm.Facility)
	assert.Equal(t, program, lm.Program)
	assert.Equal(t, msg, lm.Msg)
}
