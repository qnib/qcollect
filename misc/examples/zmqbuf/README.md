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

# ZmqBUF

A second socket provides access to a buffer of metrics which is periodically sweeped.

```
➜  QNIBCollect git:(master) ./bin/fullerite -c examples/zmqbuf/fullerite.conf                                                                                                                                                                                                                   git:(master↑5|✚1
time="28 May 16 18:27 CEST" level=info msg="Starting fullerite..." app=fullerite
time="28 May 16 18:27 CEST" level=info msg="Reading configuration file at examples/zmqbuf/fullerite.conf" app=fullerite pkg=config
time="28 May 16 18:27 CEST" level=info msg="Starting collectors..." app=fullerite
time="28 May 16 18:27 CEST" level=info msg="Reading collector configuration file at examples/zmqbuf/conf.d/Test.conf" app=fullerite pkg=config
time="28 May 16 18:27 CEST" level=info msg="Starting handlers..." app=fullerite
time="28 May 16 18:27 CEST" level=info msg="Starting handler ZmqBUF" app=fullerite
time="28 May 16 18:27 CEST" level=info msg="Created new REP socket on 'tcp://*:6060'" app=fullerite handler=ZmqBUF pkg=handler
time="28 May 16 18:27 CEST" level=info msg="Running TestCollector" app=fullerite
time="28 May 16 18:27 CEST" level=info msg="Starting to run internal metrics server on port 29090 on path /metrics" app=fullerite pkg=internalserver
time="28 May 16 18:27 CEST" level=info msg="2016-05-28 18:27:27.113396506 +0200 CEST Kicked out 0 metrics during TidyUp()" app=fullerite handler=ZmqBUF pkg=handler
time="28 May 16 18:27 CEST" level=info msg="Received request: Hello" app=fullerite handler=ZmqBUF pkg=handler
time="28 May 16 18:27 CEST" level=info msg="Received request: Hello" app=fullerite handler=ZmqBUF pkg=handler
time="28 May 16 18:27 CEST" level=info msg="2016-05-28 18:27:32.112630576 +0200 CEST Kicked out 0 metrics during TidyUp()" app=fullerite handler=ZmqBUF pkg=handler
time="28 May 16 18:27 CEST" level=info msg="Starting to emit 1 metrics" app=fullerite handler=ZmqBUF pkg=handler
time="28 May 16 18:27 CEST" level=info msg="POST of 1 metrics to ZmqBUF took 0.000085 seconds" app=fullerite handler=ZmqBUF pkg=handler
time="28 May 16 18:27 CEST" level=info msg="Received request: Hello" app=fullerite handler=ZmqBUF pkg=handler
time="28 May 16 18:27 CEST" level=info msg="2016-05-28 18:27:37.112920599 +0200 CEST Kicked out 0 metrics during TidyUp()" app=fullerite handler=ZmqBUF pkg=handler
time="28 May 16 18:27 CEST" level=info msg="Starting to emit 1 metrics" app=fullerite handler=ZmqBUF pkg=handler
time="28 May 16 18:27 CEST" level=info msg="POST of 1 metrics to ZmqBUF took 0.000110 seconds" app=fullerite handler=ZmqBUF pkg=handler
time="28 May 16 18:27 CEST" level=info msg="Received request: Hello" app=fullerite handler=ZmqBUF pkg=handler
time="28 May 16 18:27 CEST" level=info msg="2016-05-28 18:27:42.113398796 +0200 CEST Kicked out 0 metrics during TidyUp()" app=fullerite handler=ZmqBUF pkg=handler
time="28 May 16 18:27 CEST" level=info msg="Starting to emit 1 metrics" app=fullerite handler=ZmqBUF pkg=handler
time="28 May 16 18:27 CEST" level=info msg="POST of 1 metrics to ZmqBUF took 0.000126 seconds" app=fullerite handler=ZmqBUF pkg=handler
time="28 May 16 18:27 CEST" level=info msg="2016-05-28 18:27:47.113485668 +0200 CEST Kicked out 0 metrics during TidyUp()" app=fullerite handler=ZmqBUF pkg=handler
time="28 May 16 18:27 CEST" level=info msg="Starting to emit 1 metrics" app=fullerite handler=ZmqBUF pkg=handler
time="28 May 16 18:27 CEST" level=info msg="POST of 1 metrics to ZmqBUF took 0.000117 seconds" app=fullerite handler=ZmqBUF pkg=handler
time="28 May 16 18:27 CEST" level=info msg="2016-05-28 18:27:52.113961899 +0200 CEST Kicked out 0 metrics during TidyUp()" app=fullerite handler=ZmqBUF pkg=handler
time="28 May 16 18:27 CEST" level=info msg="Starting to emit 1 metrics" app=fullerite handler=ZmqBUF pkg=handler
time="28 May 16 18:27 CEST" level=info msg="POST of 1 metrics to ZmqBUF took 0.000086 seconds" app=fullerite handler=ZmqBUF pkg=handler
time="28 May 16 18:27 CEST" level=info msg="2016-05-28 18:27:57.110810578 +0200 CEST Kicked out 0 metrics during TidyUp()" app=fullerite handler=ZmqBUF pkg=handler
time="28 May 16 18:27 CEST" level=info msg="Starting to emit 1 metrics" app=fullerite handler=ZmqBUF pkg=handler
time="28 May 16 18:27 CEST" level=info msg="POST of 1 metrics to ZmqBUF took 0.000075 seconds" app=fullerite handler=ZmqBUF pkg=handler
time="28 May 16 18:28 CEST" level=info msg="2016-05-28 18:28:02.114207963 +0200 CEST Kicked out 2 metrics during TidyUp()" app=fullerite handler=ZmqBUF pkg=handler
time="28 May 16 18:28 CEST" level=info msg="Starting to emit 1 metrics" app=fullerite handler=ZmqBUF pkg=handler
time="28 May 16 18:28 CEST" level=info msg="POST of 1 metrics to ZmqBUF took 0.000185 seconds" app=fullerite handler=ZmqBUF pkg=handler
```

Within the scripts directory a little helper is provides, which REQuests the buffer and gets JSON as a RE(s)Ponse.

```
➜  zmqreq git:(match_zmqbuf) ✗ go run main.go tcp://localhost:6060 "hello.*"                                                                                                                                                                                                                git:(match_zmqbuf|✚4
2016/05/29 10:36:54 Subscriber created and connected
Sending  {"name":"hello.*","type":"gauge","dimensions":{}}
Received  {"name":"helloWorld","type":"gauge","value":0.6046602879796196,"dimensions":{"collector":"Test","testing":"yes"},"buffered":true,"time":"2016-05-29T10:36:42.231534076+02:00"}
Received  {"name":"helloWorld","type":"gauge","value":0.9405090880450124,"dimensions":{"collector":"Test","testing":"yes"},"buffered":true,"time":"2016-05-29T10:36:47.229114507+02:00"}
➜  zmqreq git:(match_zmqbuf) ✗
```

By sending a missing metric name as second argument, no response is given.

```
➜  zmqreq git:(match_zmqbuf) ✗ go run main.go tcp://localhost:6060 "FAIL.*"                                                                                                                                                                                                                 git:(match_zmqbuf|✚4
2016/05/29 10:36:59 Subscriber created and connected
Sending  {"name":"FAIL.*","type":"gauge","dimensions":{}}
➜  zmqreq git:(match_zmqbuf) ✗
```
