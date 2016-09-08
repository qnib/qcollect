# OpenTSDB Collector

The OpenTSDB collector listens for OpenTSDB line formated string on a TCP socket.


# Setup

Clone the reposiroty and make (or download) fullerite

```
$ git clone https://github.com/qnib/QNIBCollect.git
$ cd QNIBCollect
$ make (or download)
```

# Run

Run fullerite using the example configuration

```
$ ./bin/fullerite -c examples/opentsdb-collector/fullerite.conf
time="02 Jun 16 10:51 CEST" level=info msg="Starting fullerite..." app=fullerite
time="02 Jun 16 10:51 CEST" level=info msg="Reading configuration file at examples/opentsdb-collector/fullerite.conf" app=fullerite pkg=config
time="02 Jun 16 10:51 CEST" level=info msg="Starting collectors..." app=fullerite
time="02 Jun 16 10:51 CEST" level=info msg="Reading collector configuration file at examples/opentsdb-collector/conf.d/OpenTSDB.conf" app=fullerite pkg=config
time="02 Jun 16 10:51 CEST" level=info msg="Starting handlers..." app=fullerite
time="02 Jun 16 10:51 CEST" level=info msg="Starting handler Log" app=fullerite
time="02 Jun 16 10:51 CEST" level=info msg="Running OpenTSDBCollector" app=fullerite
time="02 Jun 16 10:51 CEST" level=info msg="Starting to run internal metrics server on port 29090 on path /metrics" app=fullerite pkg=internalserver
```

When emitting a metric..

```
$ echo "put sys.cpu.user host=webserver01,cpu=0 1356998400 1" |nc -w1 localhost 4242
```

... the metrics are received in fullerite.

```
time="02 Jun 16 10:51 CEST" level=info msg="Connection started: 127.0.0.1:59897" app=fullerite collector=OpenTSDB pkg=collector
time="02 Jun 16 10:51 CEST" level=warning msg="Error while reading OpenTSDB metricsEOF" app=fullerite collector=OpenTSDB pkg=collector
time="02 Jun 16 10:51 CEST" level=info msg="Connection closed: 127.0.0.1:59897" app=fullerite collector=OpenTSDB pkg=collector
time="02 Jun 16 10:51 CEST" level=info msg="Starting to emit 1 metrics" app=fullerite handler=Log pkg=handler
time="02 Jun 16 10:51 CEST" level=info msg="{\"name\":\"sys.cpu.user\",\"type\":\"gauge\",\"value\":1,\"dimensions\":{\"collector\":\"OpenTSDB\",\"cpu\":\"0\",\"host\":\"webserver01\"},\"buffered\":false,\"time\":\"2013-01-01T01:00:00+01:00\"}" app=fullerite handler=Log pkg=handler
time="02 Jun 16 10:51 CEST" level=info msg="POST of 1 metrics to Log took 0.000269 seconds" app=fullerite handler=Log pkg=handler
$
```
