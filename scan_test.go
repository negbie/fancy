package main

import (
	"bytes"
	"log"
	"log/syslog"
	"sync"
	"testing"
)

var raw = []byte("6 pad fancy {\"key1\":\"val1\", \"key2\":\"val2\"}\n")

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
	input := &Input{
		metricOnly: false,
		lineChan:   make(chan *LogLine, 1000),
	}

	stdout := bytes.NewBuffer(make([]byte, 0, 8192))
	stdin := bytes.NewBuffer(make([]byte, 0, 8192))
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		input.run(stdout, stdin)
		wg.Done()
	}()

	for i := 0; i < b.N; i++ {
		stdin.Write(raw)
	}
	wg.Wait()
}

func ping() {
	w, err := syslog.Dial("tcp", "localhost:514", syslog.LOG_DEBUG, "fancy")
	if err != nil {
		log.Fatal(err)
	}
	w.Info("ping fancy!")
}
