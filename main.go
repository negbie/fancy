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

const version = "1.3"

func main() {
	fmt.Fprintf(os.Stderr, "%v start fancy v.%s with flags %s\n", time.Now(), version, os.Args[1:])
	os.Exit(start(os.Stderr, os.Stdin, os.Args[1:]))
}

var (
	logScanNumber = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "fancy_input_scan_total",
		Help: "Total number of logs received from rsyslog fancy template"},
		[]string{"hostname", "program", "level", "tag"})
	logScanSize = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "fancy_input_raw_bytes_total",
		Help: "Total number of bytes received from rsyslog fancy template"},
		[]string{"hostname", "program"})
)

func start(stderr io.Writer, stdin io.Reader, args []string) int {
	fs := flag.NewFlagSet("fancy", flag.ContinueOnError)
	lastWarn := time.Now()
	defer fmt.Fprintf(stderr, "%v end fancy with flags %s\n", lastWarn, args)
	var (
		lokiURL    = fs.String("lokiurl", "http://localhost:3100", "Loki Server URL")
		chanSize   = fs.Int("chansize", 10000, "Loki buffered channel capacity")
		batchSize  = fs.Int("batchsize", 100*1024, "Loki will batch these bytes before sending them")
		batchWait  = fs.Int("batchwait", 4, "Loki will send logs after these seconds")
		metricOnly = fs.Bool("metriconly", false, "Only metrics for Prometheus will be exposed")
		promAddr   = fs.String("promaddr", ":9090", "Prometheus scrape endpoint address")
		promTag    = fs.String("promtag", "", "Will be used as a tag label for the fancy_input_scan_total metric")
	)

	err := ff.Parse(fs, args, ff.WithEnvVarPrefix("FANCY"))
	if err != nil {
		fmt.Fprintf(stderr, "%v ERROR: %v\n", lastWarn, err)
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
		ll, err := scanLine(s.Bytes(), *metricOnly)
		if err != nil {
			fmt.Fprintf(stderr, "%v ERROR: %v\n", lastWarn, err)
			continue
		}

		if *metricOnly {
			rawSize := float64(len(ll.Raw))
			logScanNumber.WithLabelValues(ll.Hostname, ll.Program, ll.Severity, *promTag).Inc()
			logScanSize.WithLabelValues(ll.Hostname, ll.Program).Add(rawSize)
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
