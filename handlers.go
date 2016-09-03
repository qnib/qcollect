package main

import (
	"github.com/qnib/qcollect/config"
	"github.com/qnib/qcollect/handler"
	"github.com/qnib/qcollect/metric"
)

func startHandlers(c config.Config) (handlers []handler.Handler) {
	log.Info("Starting handlers...")
	for name, config := range c.Handlers {
		handlers = append(handlers, startHandler(name, c, config))
	}
	return handlers
}

func startHandler(name string, globalConfig config.Config, instanceConfig map[string]interface{}) handler.Handler {
	log.Info("Starting handler ", name)
	handlerInst := handler.New(name)
	if handlerInst == nil {
		return nil
	}

	// apply any global configs
	handlerInst.SetInterval(config.GetAsInt(globalConfig.Interval, handler.DefaultInterval))
	handlerInst.SetPrefix(globalConfig.Prefix)
	handlerInst.SetDefaultDimensions(globalConfig.DefaultDimensions)

	// now apply the handler level configs
	handlerInst.Configure(instanceConfig)

	// now run a listener channel for each collector
	handlerInst.InitListeners(globalConfig)

	go handlerInst.Run()
	return handlerInst
}

func writeToHandlers(handlers []handler.Handler, metric metric.Metric) {
	for i := range handlers {
		handlers[i].Channel() <- metric
	}
}
