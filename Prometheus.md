# Prometheus Collector

To collect metrics from the experimental docker-1.13 endpoint `/metrics`, which provides internal metrics of the Docker-Engine in Prometheus format, a PoC Prometheus collector was created.

It reuses code from [qnib/prom2json](https://github.com/qnib/prom2json) (a fork of [prometheus/prom2json](https://github.com/prometheus/prom2json), which externalized the logic into a package).

## Example

The Docker-Engine exposes on `0.0.0.0:3376/metrics`.

```
root@swarm0:/vagrant# cat /etc/docker/daemon.json |grep metrics-addr
    "metrics-addr": "0.0.0.0:3376",
root@swarm0:/vagrant#
```

The collector fetches the metrics reusing `prom2json` code and creates metrics out of them.

The config used:

```
root@swarm0:/vagrant# cat /etc/qcollect.conf
{
    "prefix": "qcollect.",
    "interval": 5,
    "fulleritePort": 19191,
    "internalServer": {"port":"29090","path":"/metrics"},
    "collectorsConfigPath": "/etc/qcollect/conf.d",

    "collectors": [ "Prometheus" ],

    "handlers": {
      "Log": {
        "interval": "10",
        "max_buffer_size": "100"
      }
    }
}
root@swarm0:/vagrant# cat /etc/qcollect/conf.d/Prometheus.conf
{}
root@swarm0:/vagrant#
```

The output looks like this:

```
root@swarm0:/vagrant# ./qcollect
time="2016-11-20T11:26:12.825730442Z" level=info msg="Starting qcollect..." app=qcollect
time="2016-11-20T11:26:12.827011083Z" level=info msg="Reading configuration file at /etc/qcollect.conf" app=qcollect pkg=config
time="2016-11-20T11:26:12.829439145Z" level=info msg="Starting collectors..." app=qcollect
time="2016-11-20T11:26:12.830266898Z" level=info msg="Reading collector configuration file at /etc/qcollect/conf.d/Prometheus.conf" app=qcollect pkg=config
time="2016-11-20T11:26:12.831414147Z" level=info msg="Starting handlers..." app=qcollect
time="2016-11-20T11:26:12.83198108Z" level=info msg="Starting handler Log" app=qcollect
time="2016-11-20T11:26:12.833991792Z" level=info msg="Running PrometheusCollector" app=qcollect
time="2016-11-20T11:26:12.835942599Z" level=info msg="Starting to run internal metrics server on port 29090 on path /metrics" app=qcollect pkg=internalserver
time="2016-11-20T11:26:17.852366215Z" level=info msg="Dunno what to do with 'HISTOGRAM'" app=qcollect collector=Prometheus pkg=collector
time="2016-11-20T11:26:17.8527669Z" level=info msg="Dunno what to do with 'HISTOGRAM'" app=qcollect collector=Prometheus pkg=collector
time="2016-11-20T11:26:17.853093419Z" level=info msg="Dunno what to do with 'HISTOGRAM'" app=qcollect collector=Prometheus pkg=collector
time="2016-11-20T11:26:17.853569184Z" level=info msg="Dunno what to do with 'HISTOGRAM'" app=qcollect collector=Prometheus pkg=collector
time="2016-11-20T11:26:22.836683947Z" level=info msg="Starting to emit 60 metrics" app=qcollect handler=Log pkg=handler
time="2016-11-20T11:26:22.838371066Z" level=info msg="{\"name\":\"engine_daemon_engine_cpus_cpus\",\"type\":\"GAUGE\",\"value\":1,\"dimensions\":{\"collector\":\"Prometheus\"},\"buffered\":false,\"time\":\"2016-11-20T11:26:17.852700975Z\"}" app=qcollect handler=Log pkg=handler
time="2016-11-20T11:26:22.838404541Z" level=info msg="{\"name\":\"engine_daemon_engine_info\",\"type\":\"GAUGE\",\"value\":1,\"dimensions\":{\"collector\":\"Prometheus\"},\"buffered\":false,\"time\":\"2016-11-20T11:26:17.852706458Z\"}" app=qcollect handler=Log pkg=handler
time="2016-11-20T11:26:22.838418831Z" level=info msg="{\"name\":\"engine_daemon_engine_memory_bytes\",\"type\":\"GAUGE\",\"value\":2.097631232e+09,\"dimensions\":{\"collector\":\"Prometheus\"},\"buffered\":false,\"time\":\"2016-11-20T11:26:17.852717757Z\"}" app=qcollect handler=Log pkg=handler
time="2016-11-20T11:26:22.83843021Z" level=info msg="{\"name\":\"engine_daemon_events_subscribers_total\",\"type\":\"GAUGE\",\"value\":0,\"dimensions\":{\"collector\":\"Prometheus\"},\"buffered\":false,\"time\":\"2016-11-20T11:26:17.852719931Z\"}" app=qcollect handler=Log pkg=handler
*snip*
```

## TODO

As seen above, histograms are not yet integrated...
[https://github.com/qnib/qcollect/blob/0.7.0.0/collector/prometheus.go#L102-L104](https://github.com/qnib/qcollect/blob/0.7.0.0/collector/prometheus.go#L102-L104)

```
        /*} else if f.Type == "HISTOGRAM" {
            //create histogram metrics?
            continue
        */
        } else {
            p.log.Debugf("Dunno what to do with '%s'", f.Type)
            continue
}
```
```
