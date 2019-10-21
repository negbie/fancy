package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/peterbourgon/ff"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const version = "1.1"

func main() {
	fmt.Fprintf(os.Stderr, "%v start fancy v.%s with %s\n", time.Now(), version, os.Args[1:])
	os.Exit(parseAndRun(os.Stderr, os.Stdin, os.Args[1:]))
}

var (
	logScanNumber = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "fancy_input_scan_total",
		Help: "Total number of logs received from rsyslog fancy template"},
		[]string{"hostname", "program", "level", "facility"})
	logHostProgSize = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "fancy_input_raw_bytes_total",
		Help: "Total number of bytes received from rsyslog fancy template"},
		[]string{"hostname", "program"})
)

func parseAndRun(stderr io.Writer, stdin io.Reader, args []string) int {
	fs := flag.NewFlagSet("fancy", flag.ContinueOnError)
	var (
		lokiURL    = fs.String("lokiurl", "http://localhost:3100", "Loki Server URL")
		chanSize   = fs.Int("chansize", 10000, "Loki buffered channel capacity")
		batchSize  = fs.Int("batchsize", 100*1024, "Loki will batch some bytes before sending them")
		batchWait  = fs.Int("batchwait", 4, "Loki will send logs after some seconds")
		metricOnly = fs.Bool("metriconly", false, "Only metrics for prometheus will be exposed")
		promAddr   = fs.String("promaddr", ":9090", "Metrics endpoint address")
	)

	lastWarn := time.Now()
	err := ff.Parse(fs, args, ff.WithEnvVarPrefix("FANCY"))
	if err != nil {
		if err != flag.ErrHelp {
			fmt.Fprintf(stderr, "%v ERROR: %v\n", lastWarn, err)
		}
		return 1
	}

	var lineChan chan *LogLine
	if *metricOnly {
		go func() {
			http.Handle("/metrics", promhttp.Handler())
			err := http.ListenAndServe(*promAddr, nil)
			if err != nil {
				fmt.Fprintf(stderr, "%v ERROR: %v\n", lastWarn, err)
			}
		}()
	} else {
		lineChan = make(chan *LogLine, *chanSize)
		l, err := NewLoki(lineChan, *lokiURL, *batchSize, *batchWait)
		if err != nil {
			fmt.Fprintf(stderr, "%v ERROR: %v\n", lastWarn, err)
		}
		go l.Run()
	}

	s := bufio.NewScanner(stdin)
	for s.Scan() {
		ll, err := scanLine(s.Bytes())
		if err != nil {
			return 1
		}

		if *metricOnly {
			logSize := float64(len(ll.Raw))
			logScanNumber.WithLabelValues(ll.Hostname, ll.Program, ll.Severity, ll.Facility).Inc()
			logHostProgSize.WithLabelValues(ll.Hostname, ll.Program).Add(logSize)
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
