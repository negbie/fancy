package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const version = "1.5"
const scanSize = 16

func main() {
	fs := flag.NewFlagSet("fancy", flag.ExitOnError)
	var (
		cmd           = fs.String("cmd", "", "Send input msg to external command and use it's output as new msg")
		lokiURL       = fs.String("loki-url", "http://localhost:3100", "Loki Server URL")
		lokiChanSize  = fs.Int("loki-chan-size", 10000, "Loki buffered channel capacity")
		lokiBatchSize = fs.Int("loki-batch-size", 100*1024, "Loki will batch these bytes before sending them")
		lokiBatchWait = fs.Int("loki-batch-wait", 4, "Loki will send logs after these seconds")
		promOnly      = fs.Bool("prom-only", false, "Only metrics for Prometheus will be exposed")
		promAddr      = fs.String("prom-addr", ":9090", "Prometheus scrape endpoint address")
		promTag       = fs.String("prom-tag", "", "Will be used as a static label for the fancy_input_scan_total metric")
		promTagFilter = fs.String("prom-tag-filter", "", "Use prom-tag only when msg contains this string")
	)
	fs.Parse(os.Args[1:])

	t := time.Now()
	defer fmt.Fprintf(os.Stderr, "%v end fancy with flags %s\n", t, os.Args[1:])

	input := &Input{
		cmd:           strings.Fields(*cmd),
		promOnly:      *promOnly,
		promTag:       *promTag,
		promTagFilter: []byte(*promTagFilter),
		scanChan:      make(chan [scanSize][]byte, 1000),
	}

	if *promOnly {
		go func() {
			http.Handle("/metrics", promhttp.Handler())
			err := http.ListenAndServe(*promAddr, nil)
			if err != nil {
				fmt.Fprintf(os.Stderr, "%v ERROR: %v\n", t, err)
			}
		}()
	} else if len(*lokiURL) > 3 {
		input.useLoki = true
		input.lineChan = make(chan *LogLine, *lokiChanSize)
		l, err := NewLoki(input.lineChan, *lokiURL, *lokiBatchSize, *lokiBatchWait)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v ERROR: %v\n", t, err)
		}
		go l.Run()
	}

	fmt.Fprintf(os.Stderr, "%v run fancy v.%s with flags %s\n", time.Now(), version, os.Args[1:])
	for i := 0; i < 8; i++ {
		go input.process()
	}

	input.scan(os.Stderr, os.Stdin)
	os.Exit(0)
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

type Input struct {
	cmd           []string
	cache         Cache
	useLoki       bool
	scanChan      chan [scanSize][]byte
	lineChan      chan *LogLine
	promOnly      bool
	promTag       string
	promTagFilter []byte
}

type Cache struct {
	buf [scanSize][]byte
	pos int
}

func batchScan(c chan [scanSize][]byte, cache *Cache, value []byte) {
	cache.buf[cache.pos] = value
	cache.pos++
	if cache.pos == scanSize {
		c <- cache.buf
		cache.pos = 0
	}
}

func (in *Input) scan(stderr io.Writer, stdin io.Reader) {
	var err error
	r := bufio.NewReader(stdin)
	line := make([]byte, 0, 8192)
	defer close(in.scanChan)
	for {
		line, err = r.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				fmt.Fprintf(stderr, "%v INFO: %v\n", time.Now(), err)
				break
			}
			fmt.Fprintf(stderr, "%v ERROR: %v\n", time.Now(), err)
			break
		}
		batchScan(in.scanChan, &in.cache, line)
	}
}

func (in *Input) process() {
	t := time.Now()
	promTag := in.promTag
	for s := range in.scanChan {
		for i := 0; i < len(s); i++ {
			ll, err := parseLine(s[i], in.promOnly)
			if err != nil {
				fmt.Fprintf(os.Stderr, "%v ERROR: %v\n", time.Now(), err)
				continue
			}

			if in.promOnly {
				if len(in.promTagFilter) > 0 {
					promTag = ""
					if bytes.Contains(ll.Raw[ll.MsgPos:], in.promTagFilter) {
						promTag = in.promTag
					}
				}

				rawSize := float64(len(ll.Raw))
				logScanNumber.WithLabelValues(ll.Hostname, ll.Program, ll.Severity, promTag).Inc()
				logScanSize.WithLabelValues(ll.Hostname, ll.Program).Add(rawSize)
				continue
			}

			if len(in.cmd) > 0 && in.useLoki {
				c := exec.Command(in.cmd[0], in.cmd[1:]...)
				c.Stdin = bytes.NewReader(ll.Raw[ll.MsgPos:])
				out, err := c.Output()
				if err != nil {
					fmt.Fprintf(os.Stderr, "%v ERROR: %v\n", time.Now(), err)
					continue
				}
				ll.Msg = string(out)
			}

			if in.useLoki {
				select {
				case in.lineChan <- ll:
				default:
					if time.Since(t) > 1e9 {
						fmt.Fprintf(os.Stderr, "%v ERROR: overflowing Loki buffered channel capacity\n", t)
					}
					t = time.Now()
				}
			}
		}
	}
}
