# Diamond Collector

To run the diamond collector a twofold approach is needed.

- run qcollect with the `Diamond` collector
- run the `server.py`

First change to the repository:

```
$ git clone git@github.com:qnib/qcollect
$ cd qcollect
```

Run qcollect using the example configuration

```
$ ./bin/qcollect -c ./resources/examples/diamond-collector/qcollect.conf
time="2016-09-25T12:20:51.869901866+02:00" level=info msg="Starting qcollect..." app=qcollect
time="2016-09-25T12:20:51.87002892+02:00" level=info msg="Reading configuration file at resources/examples/diamond-collector/qcollect.conf" app=qcollect pkg=config
time="2016-09-25T12:20:51.870238746+02:00" level=info msg="Starting collectors..." app=qcollect
time="2016-09-25T12:20:51.87026336+02:00" level=info msg="Reading collector configuration file at ./resources/examples/diamond-collector/conf.d/Diamond.conf" app=qcollect pkg=config
time="2016-09-25T12:20:51.870323756+02:00" level=info msg="Starting handlers..." app=qcollect
time="2016-09-25T12:20:51.870340062+02:00" level=info msg="Starting handler Log" app=qcollect
time="2016-09-25T12:20:51.870475527+02:00" level=info msg="Running DiamondCollector" app=qcollect
time="2016-09-25T12:20:51.870665531+02:00" level=info msg="Starting to run internal metrics server on port 29090 on path /metrics" app=qcollect pkg=internalserver```
and start the `server.py`:

```
$ python resources/diamond/server.py -c ./resources/examples/diamond-collector/qcollect.conf
['./resources/diamond/collectors']
```

Now the `server.py` will connect to the qcollect daemon and start sending metrics.

```
time="2016-09-25T12:21:40.961949851+02:00" level=info msg="Connection started: 127.0.0.1:57742" app=qcollect collector=Diamond pkg=collector
time="2016-09-25T12:21:41.876172737+02:00" level=info msg="Starting to emit 20 metrics" app=qcollect handler=Log pkg=handler
time="2016-09-25T12:21:41.876452709+02:00" level=info msg="{\"name\":\"cpu.user\",\"type\":\"cumcounter\",\"value\":30581.22,\"dimensions\":{\"collector\":\"cpu\",\"core\":\"0\",\"diamond\":\"yes\",\"prefix\":\"servers\"},\"buffered\":false,\"time\":\"2016-09-25T12:21:40.962764049+02:00\"}" app=qcollect handler=Log pkg=handler
...
```

### Important Config

The following configuration files are needed, even though they just contain `{}`.

- `conf.d/Diamond.conf`
- `conf.d/CPUCollector.conf`
