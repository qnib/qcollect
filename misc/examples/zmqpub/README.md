# ZmqPUB

Simple ZMQ socket which publishes the metrics as JSON

```
$ ./bin/fullerite -c examples/zmqpub/fullerite.conf
time="20 May 16 14:33 CEST" level=info msg="Starting fullerite..." app=fullerite
time="20 May 16 14:33 CEST" level=info msg="Reading configuration file at examples/zmqpub/fullerite.conf" app=fullerite pkg=config
time="20 May 16 14:33 CEST" level=info msg="Starting collectors..." app=fullerite
time="20 May 16 14:33 CEST" level=info msg="Reading collector configuration file at examples/zmqpub/conf.d/Test.conf" app=fullerite pkg=config
time="20 May 16 14:33 CEST" level=info msg="Starting handlers..." app=fullerite
time="20 May 16 14:33 CEST" level=info msg="Starting handler Log" app=fullerite
time="20 May 16 14:33 CEST" level=info msg="Starting handler ZmqPUB" app=fullerite
time="20 May 16 14:33 CEST" level=info msg="Running TestCollector" app=fullerite
time="20 May 16 14:33 CEST" level=info msg="Created new PUB socket on 'tcp://*:5555'" app=fullerite handler=ZmqPUB pkg=handler
time="20 May 16 14:33 CEST" level=info msg="Starting to run internal metrics server on port 29090 on path /metrics" app=fullerite pkg=internalserver
time="20 May 16 14:33 CEST" level=info msg="Starting to emit 1 metrics" app=fullerite handler=Log pkg=handler
time="20 May 16 14:33 CEST" level=info msg="{\"name\":\"helloWorld\",\"type\":\"gauge\",\"value\":0.6046602879796196,\"dimensions\":{\"collector\":\"Test\",\"testing\":\"yes\"},\"time\":\"2016-05-20T14:33:54.922951232+02:00\"}" app=fullerite handler=Log pkg=handler
time="20 May 16 14:33 CEST" level=info msg="Starting to emit 1 metrics" app=fullerite handler=ZmqPUB pkg=handler
time="20 May 16 14:33 CEST" level=info msg="POST of 1 metrics to Log took 0.000243 seconds" app=fullerite handler=Log pkg=handler
time="20 May 16 14:33 CEST" level=info msg="POST of 1 metrics to ZmqPUB took 0.000379 seconds" app=fullerite handler=ZmqPUB pkg=handler
time="20 May 16 14:34 CEST" level=info msg="Starting to emit 1 metrics" app=fullerite handler=ZmqPUB pkg=handler
time="20 May 16 14:34 CEST" level=info msg="POST of 1 metrics to ZmqPUB took 0.000139 seconds" app=fullerite handler=ZmqPUB pkg=handler
time="20 May 16 14:34 CEST" level=info msg="Starting to emit 2 metrics" app=fullerite handler=Log pkg=handler
time="20 May 16 14:34 CEST" level=info msg="{\"name\":\"helloWorld\",\"type\":\"gauge\",\"value\":0.9405090880450124,\"dimensions\":{\"collector\":\"Test\",\"testing\":\"yes\"},\"time\":\"2016-05-20T14:33:59.924125235+02:00\"}" app=fullerite handler=Log pkg=handler
time="20 May 16 14:34 CEST" level=info msg="{\"name\":\"helloWorld\",\"type\":\"gauge\",\"value\":0.6645600532184904,\"dimensions\":{\"collector\":\"Test\",\"testing\":\"yes\"},\"time\":\"2016-05-20T14:34:04.922250183+02:00\"}" app=fullerite handler=Log pkg=handler
time="20 May 16 14:34 CEST" level=info msg="POST of 2 metrics to Log took 0.000150 seconds" app=fullerite handler=Log pkg=handler
time="20 May 16 14:34 CEST" level=info msg="Starting to emit 1 metrics" app=fullerite handler=ZmqPUB pkg=handler
time="20 May 16 14:34 CEST" level=info msg="POST of 1 metrics to ZmqPUB took 0.000210 seconds" app=fullerite handler=ZmqPUB pkg=handler
```

Within the script section a little client is provided.

```
$ go run main.go tcp://localhost:5555
2016/05/20 14:33:53 Subscriber created and connected
2016/05/20 14:33:59 Message '{"name":"helloWorld","type":"gauge","value":0.6046602879796196,"dimensions":{"collector":"Test","testing":"yes"},"time":"2016-05-20T14:33:54.922951232+02:00"}' received
2016/05/20 14:34:04 Message '{"name":"helloWorld","type":"gauge","value":0.9405090880450124,"dimensions":{"collector":"Test","testing":"yes"},"time":"2016-05-20T14:33:59.924125235+02:00"}' received
2016/05/20 14:34:09 Message '{"name":"helloWorld","type":"gauge","value":0.6645600532184904,"dimensions":{"collector":"Test","testing":"yes"},"time":"2016-05-20T14:34:04.922250183+02:00"}' received
```
