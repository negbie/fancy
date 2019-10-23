package main

import (
	"log"
	"log/syslog"
	"testing"
)

var raw = []byte("6 pad fancy {\"key1\":\"val1\", \"key2\":\"val2\"}")

type TestCase struct {
	input []byte
	want  string
	err   error
}

func Test_scanLine(t *testing.T) {
	cases := []TestCase{
		TestCase{
			input: []byte("6 pad fancy {\"key1\":\"val1\", \"key2\":\"val2\"}"),
			want:  "6 pad fancy {\"key1\":\"val1\", \"key2\":\"val2\"}",
			err:   nil,
		},
		TestCase{
			input: []byte("6 pad fancy"),
			want:  "",
			err:   errLength,
		},
		TestCase{
			input: []byte("9 pad fancy {\"key1\":\"val1\", \"key2\":\"val2\"}"),
			want:  "",
			err:   errLevel,
		},
		TestCase{
			input: []byte("6padfancy {\"key1\":\"val1\", \"key2\":\"val2\"}"),
			want:  "",
			err:   errTemplate,
		},
	}

	for _, c := range cases {
		got, err := scanLine(c.input, false)
		if err != c.err || got.String() != c.want {
			t.Errorf("got %q,%v but want %q,%v", got.String(), err, c.want, c.err)
		}
	}
}

func Benchmark_scanLine(b *testing.B) {
	metricOnly := true
	for i := 0; i < b.N; i++ {
		scanLine(raw, metricOnly)
	}
}

func ping() {
	w, err := syslog.Dial("tcp", "localhost:514", syslog.LOG_DEBUG, "fancy")
	if err != nil {
		log.Fatal(err)
	}
	w.Info("ping fancy!")
}
