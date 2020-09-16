package main

import (
	"bytes"
	"log"
	"log/syslog"
	"sync"
	"testing"
)

var raw = []byte("2019-10-29T16:21:22.230666+01:00 6 pad fancy {\"key1\":\"val1\", \"key2\":\"val2\"}\n")

type TestCase struct {
	input []byte
	want  string
	err   error
}

func Test_parseLine(t *testing.T) {
	cases := []TestCase{
		{
			input: []byte("2019-10-29T16:21:22.230666+01:00 6 pad fancy {\"key1\":\"val1\", \"key2\":\"val2\"}"),
			want:  "6 pad fancy {\"key1\":\"val1\", \"key2\":\"val2\"}",
			err:   nil,
		},
		{
			input: []byte("2019-10-29T16:21:22.230666+01:00 6 pad fancy"),
			want:  "",
			err:   errLength,
		},
		{
			input: []byte("2019-10-29T16:21:22.230666+01:00 9 pad fancy {\"key1\":\"val1\", \"key2\":\"val2\"}"),
			want:  "",
			err:   errLevel,
		},
	}

	for _, c := range cases {
		got, err := parseLine(c.input, false)
		if err != c.err || got.String() != c.want {
			t.Errorf("\nrecv %q,%s\nwant %q,%s\n", got.String(), err, c.want, c.err)
		}
	}
}

func Benchmark_parseLine(b *testing.B) {
	input := &Input{
		//cmd:        []string{"tr", "[a-z]", "[A-Z]"},
		promOnly:        true,
		staticTagFilter: []byte("val1"),
		lineChan:        make(chan *LogLine, 1000),
		scanChan:        make(chan [scanSize][]byte, 1000),
	}

	var stdout bytes.Buffer
	var stdin bytes.Buffer

	for i := 0; i < b.N; i++ {
		stdin.Write(raw)
	}

	for i := 0; i < 8; i++ {
		go input.process()
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		input.scan(&stdout, &stdin)
		wg.Done()
	}()
	wg.Wait()
}

func ping() {
	w, err := syslog.Dial("tcp", "localhost:514", syslog.LOG_DEBUG, "fancy")
	if err != nil {
		log.Fatal(err)
	}
	w.Info("ping fancy!")
}
