# qcollect
Diamond compatible metrics collector. Based on yelp/fullerite fork.

# Old README of QNIBCollect
**'Why fork?!'** you ask? I want to prototype my thoughts about a metrics collection framework and to not interfere with what Yelp wants to drive the project into I thought I should fork and see where I can go.

Of course keep in touch with upstream.

## Vision

So 'butter bei die Fisch'; what do I want to do what makes it necessary to fork?

#### Event vs. Process Time (DONE [#4](https://github.com/qnib/QNIBCollect/pull/4))

For now fullerite collects metrics without adding a timestamp. The timestamp is added to the metric by the handler, hence when the handler consumes the metrics channel.

This might be reasonable if the handlers are consuming the metrics right away. Which segways to the next ideas...

#### High precision collection

To not miss important samples I would like to sample at hight frequency no matter what. OK, it should be the minimum frequency feasible for the task at hand.

#### Bulky Handlers (DONE with Event time)

Most back-ends like the notion of getting the metrics thrown at them in bulks. As the metrics within fullerite are presented to the handlers as list, it is not far stretched to collect the metrics at hight precision (1s) and only send it every once so often (30s). For this to work a metric is needed, which is timestamped during collection.

With hight precision collectors this would pollute the back-ends, that is why I need...

#### Downsampling

The handler should be able to downsample the metrics. If he gets metrics which are sampled at 1 second interval, he should take the metrics and aggregate over each metric-key to fit the interval he is handling the metrics (e.g. 30s).

#### Aggregation Layer

Even better, all metrics are send to an aggregation layer which could aggregate in different intervals and push the metrics to an aggregate-channel. The default aggregator would directly forward the metrics to the `0s` aggregation channel (real-time).

##### Multiple Handler of the same kind

AFAIK the current fullerite is not able to have two handlers of the same kind.
I would like to establish that to allow sending `real-time` to a hot cache and `aggregated` to a more long-time storage backend.

##### Websocket handler

As to be able to do stuff like Netflix's Vector does, I would like to expose the metrics on a web socket.

##### ZMQ colector/handler (collector: [#16](https://github.com/qnib/QNIBCollect/pull/16))

To allow a hierarchy of collectors (node, rack, DC, ...) a ZMQ PUB/SUB socket would be nice. Maybe the web socket would be sufficient to make this happen...

## fullerite

*Fullerite is a metrics collection tool*. It is different than other collection tools (e.g. diamond, collectd) in that it supports multidimensional metrics from its core. It is also meant to innately support easy concurrency. Collectors and handler are sufficiently isolated to avoid having one misbehaving component affect the rest of the system. Generally, an instance of fullerite runs as a daemon on a box collecting the configured metrics and reports them via different handlers to endpoints such as graphite, kairosdb, signalfx, or datadog.

A summary of interesting features of fullerite include:
 * Fully compatible with diamond collectors
 * Written in Go for easy reliable concurrency
 * Configurable set of handlers and collectors
 * Native support for dimensionalized metrics
 * Internal metrics to track handler performance

Fullerite is also able to run [Diamond](https://github.com/python-diamond/Diamond) collectors natively. This means you don't need to port your python code over to Go. We'll do the heavy lifting for you.

### success story
  * Running on 1,000s of machines
  * Running on AWS and real hardware all over the world
  * Running 8-12 collectors and 1-2 handlers at the same time
  * Emitting over 5,000 metrics per flush interval on average per box
  * Well over 10 million metrics per minute

### how it works
Fullerite works by spawning a separate goroutines for each collector and handler then acting as the conduit between the two. Each collector and handler can be individually configured with a nested JSON map in the configuration. But sane defaults are provided.

The `fullerite_diamond_server` is a process that starts each diamond collector in python as a separate process. The listening collector in go must also be configured on. Doing this each diamond collector will connect to the server and then start piping metrics to the collector. The server handles the transient connections and other such issues by spawning a new goroutine for each of the connecting collectors.

![Alt text](/fullerite_arch.jpg?raw=true "Optional Title")

### using fullerite
Fullerite makes a deb package that can be installed onto a linux box. It has been tested a lot with Ubuntu trusty, lucid, and precise. Once installed it can be controlled like any normal service:

    $ service fullerite [status | start | stop]
    $ service fullerite_diamond_server [status | start | stop]

By default it logs out to `/var/log/fullerite/*`. It runs as user `fuller`. This can all be changed by editing the `/etc/default/fullerite.conf` file. See the upstart scripts for [fullerite](deb/etc/init/fullerite) and [fullerite_diamond_server](deb/etc/init/fullerite_diamond_server) for more info.

You can also run fullerite directly using the commands: `run-fullerite.sh` and `run-diamond-collectors.sh`. These both have command line args that are good to use.

Finally, fullerite is just a simple go binary. You can manually invoke it and pass it arguments as you'd like.

# Contributing to fullerite

We welcome all contribution to fullerite, If you have a feature request or you want to improve
existing functionality of fullerite - it is probably best to open a pull request with your changes.

## Adding new dependency

If you want to add new external dependency to fullerite, please make sure it is added to `Gomfile`.
Do not forget to specify `TAG` or `commit_id` of external git repository.  More information about
`Gomfile` can be found at https://github.com/mattn/gom.

## Ensure code is formatted, tested and passes golint.

Running `make` should do all of the above. If you see any failures or errors while running `make`,
please fix them before opening a pull request.

## Building and compiling

Running `make` should build the fullerite go binary and place it in the `bin` directory.

## Building package fails or gom install fails

If you have vendored external dependencies in `src/` directory or `pkg` directory from old build configuration, you should
delete `src/github.com`, `pkg` and `src/golang.org` before running `gom install` or attempting to build the package.

Aforementioned directories are artifacts of old build configuration before we moved to using `gom` for managing dependencies.
