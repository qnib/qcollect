# Diamond Collector

To run the diamond collector a twofold approach is needed.

- run fullerite with the `Diamond` collector
- run the `server.py`

First change to the repository:

```
$ git clone git@github.com:Yelp/fullerite
$ cd fullerite
```

Run fullerite using the example configuration

```
$ ./bin/fullerite -c examples/diamond-collector/fullerite.conf
time="04 May 16 17:17 CEST" level=info msg="Starting fullerite..." app=fullerite
time="04 May 16 17:17 CEST" level=info msg="Reading configuration file at examples/diamond-collector/fullerite.conf" app=fullerite pkg=config
time="04 May 16 17:17 CEST" level=info msg="Starting collectors..." app=fullerite
time="04 May 16 17:17 CEST" level=info msg="Reading collector configuration file at examples/diamond-collector/conf.d/Diamond.conf" app=fullerite pkg=config
time="04 May 16 17:17 CEST" level=info msg="Starting handlers..." app=fullerite
time="04 May 16 17:17 CEST" level=info msg="Starting handler Log" app=fullerite
time="04 May 16 17:17 CEST" level=info msg="Running DiamondCollector" app=fullerite
time="04 May 16 17:17 CEST" level=info msg="Starting to run internal metrics server on port 29090 on path /metrics" app=fullerite pkg=internalserver
```
and start the `server.py`:

```
$ python src/diamond/server.py -c examples/diamond-collector/fullerite.conf
```

Now the `server.py` will connect to the fullerite daemon and start sending metrics.

```
time="04 May 16 17:17 CEST" level=info msg="Connection started: 127.0.0.1:52830" app=fullerite collector=Diamond pkg=collector
time="04 May 16 17:17 CEST" level=info msg="Starting to emit 20 metrics" app=fullerite handler=Log pkg=handler
time="04 May 16 17:17 CEST" level=info msg="{\"name\":\"cpu.user\",\"type\":\"cumcounter\",\"value\":58276.53,\"dimensions\":{\"collector\":\"cpu\",\"core\":\"0\",\"diamond\":\"yes\",\"prefix\":\"servers\"}}" app=fullerite handler=Log pkg=handler
time="04 May 16 17:17 CEST" level=info msg="{\"name\":\"cpu.nice\",\"type\":\"cumcounter\",\"value\":0,\"dimensions\":{\"collector\":\"cpu\",\"core\":\"0\",\"diamond\":\"yes\",\"prefix\":\"servers\"}}" app=fullerite handler=Log pkg=handler
...
```

### Important Config

The following configuration files are needed, even though they just contain `{}`.

- `conf.d/Diamond.conf`
- `conf.d/CPUCollector.conf`
 