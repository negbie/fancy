package main

import (
	"bytes"
	"log"
	"log/syslog"
	"testing"

	"github.com/stretchr/testify/assert"
)

var raw = []byte("6 pad kern fancy {\"key1\":\"val1\", \"key2\":\"val2\"}")

func Test_scanLine(t *testing.T) {
	severity := "info"
	hostname := "pad"
	facility := "kern"
	program := "fancy"
	msg := "{\"key1\":\"val1\", \"key2\":\"val2\"}"
	res := bytes.NewBufferString("")

	scanLine(raw)
	assert.Equal(t, 0, res.Len())

	res.Reset()
	ll, err := scanLine(raw)
	if err != nil || !ll.Valid() {
		t.Fail()
	}
	assert.Equal(t, severity, ll.Severity)
	assert.Equal(t, hostname, ll.Hostname)
	assert.Equal(t, facility, ll.Facility)
	assert.Equal(t, program, ll.Program)
	assert.Equal(t, msg, ll.Msg)
}

func Benchmark_scanLine(b *testing.B) {
	for i := 0; i < b.N; i++ {
		scanLine(raw)
	}
}

func ping() {
	w, err := syslog.Dial("tcp", "localhost:514", syslog.LOG_DEBUG, "fancy")
	if err != nil {
		log.Fatal(err)
	}
	w.Info("ping fancy!")
}
