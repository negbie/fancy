<img src="https://user-images.githubusercontent.com/20154956/67162203-297e5300-f362-11e9-899b-4644d3084a02.png" width="200" height="160">

**fancy** let's you fanout `rsyslog` to [Loki](https://github.com/grafana/loki) and is meant to be executed by `rsyslog` under
[omprog](http://www.rsyslog.com/doc/master/configuration/modules/omprog.html)

## Usage

```
➜  fancy git:(master) ✗ ./fancy -h
Usage of fancy:
  -cmd string
        Send input msg to external command and use it's output as new msg
  -loki-batch-size int
        Loki will batch these bytes before sending them (default 1048576)
  -loki-batch-wait int
        Loki will send logs after these seconds (default 4)
  -loki-chan-size int
        Loki buffered channel capacity (default 10000)
  -loki-url string
        Loki Server URL (default "http://localhost:3100")
  -prom-addr string
        Prometheus scrape endpoint address (default ":9090")
  -prom-only
        Only metrics for Prometheus will be exposed
  -environment string
        Set an environment tag
  -service string
        Set a service tag
  -static-tag string
        Will be used as a static label value with the name static_tag
  -static-tag-filter string
        Set static-tag only when msg contains this string
```
## Setup

1. Download **fancy** from [releases](https://github.com/negbie/fancy/releases) or compile it from sources (`go build .`)
2. Make **fancy** executable using `chmod +x fancy`.
3. Move **fancy** to /opt with `mv fancy /opt/`.
4. Edit rsyslog config with `vim /etc/rsyslog.conf`.

```bash
    module(load="omprog")

    template(
        name="LokiFormat"
        type="string"
        string="%timegenerated:::date-rfc3339% %syslogseverity% %hostname% %programname%%msg%\n"
        )

    action(
        type="omprog"
        name="loki"
        template="LokiFormat"
        binary="/opt/fancy --environment dev --service example_service --loki-url https://your_endpoint/api/prom/push"
        )
```
5. Make sure you have set the right Loki URL.
6. Restart `rsyslog` using `systemctl restart rsyslog`.
7. Profit.