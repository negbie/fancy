package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/peterbourgon/ff"
)

const version = "1.0"

func main() {
	fmt.Fprintf(os.Stderr, "%v start fancy version %s\n", time.Now(), version)
	os.Exit(parseAndRun(os.Stderr, os.Stdin, os.Args[1:]))
}

func parseAndRun(stderr io.Writer, stdin io.Reader, args []string) int {
	fs := flag.NewFlagSet("fancy", flag.ContinueOnError)
	var (
		lokiURL    = fs.String("lokiurl", "http://localhost:3100", "Loki Server URL")
		chanSize   = fs.Int("chansize", 20000, "Loki buffered channel capacity")
		batchSize  = fs.Int("batchsize", 600*1024, "Loki will batch some bytes before sending them")
		batchWait  = fs.Int("batchwait", 4, "Loki will send logs after some seconds")
		metricOnly = fs.Bool("metriconly", false, "Only metrics for prometheus will be exposed")
	)

	err := ff.Parse(fs, args, ff.WithEnvVarPrefix("FANCY"))
	if err != nil {
		if err != flag.ErrHelp {
			fs.Output().Write([]byte(fmt.Sprintf("\n%s\n", err)))
		}
		return 1
	}

	lastWarn := time.Now()
	lineChan := make(chan *LogLine, *chanSize)
	l, err := NewLoki(lineChan, *lokiURL, *batchSize, *batchWait)
	if err != nil {
		fmt.Fprintf(stderr, "%v ERROR: %v\n", lastWarn, err)
	}
	go l.Run()

	s := bufio.NewScanner(stdin)
	for s.Scan() {
		ll, err := scanLine(s.Bytes())
		if err != nil {
			return 1
		}

		if *metricOnly {
			continue
		}

		select {
		case lineChan <- ll:
		default:
			if time.Since(lastWarn) > 1e9 {
				fmt.Fprintf(stderr, "%v ERROR: overflowing Loki buffered channel capacity\n", lastWarn)
			}
			lastWarn = time.Now()
		}
	}
	return 0
}
