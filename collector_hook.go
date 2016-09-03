package main

import (
	"fmt"
	"os"

	"github.com/qnib/qcollect/handler"
	"github.com/qnib/qcollect/metric"

	"github.com/Sirupsen/logrus"
)

// LogErrorHook to send errors via handlers.
type LogErrorHook struct {
	handlers []handler.Handler

	// intentionally exported
	log *logrus.Entry
}

// NewLogErrorHook creates a hook to be added to the collector logger
// so that errors are forwarded as a metric to the handlers.
func NewLogErrorHook(handlers []handler.Handler) *LogErrorHook {
	hookLog := log.WithFields(logrus.Fields{"hook": "LogErrorHook"})
	return &LogErrorHook{handlers, hookLog}
}

// Fire action to take when log is fired.
func (hook *LogErrorHook) Fire(entry *logrus.Entry) error {
	_, err := entry.String()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to read entry, %v", err)
		return err
	}

	go hook.reportErrors(entry)
	return nil
}

// Levels covered by this hook
func (hook *LogErrorHook) Levels() []logrus.Level {
	return []logrus.Level{
		logrus.PanicLevel,
		logrus.FatalLevel,
		logrus.ErrorLevel,
	}
}

func (hook *LogErrorHook) reportErrors(entry *logrus.Entry) {
	metric := metric.New("qcollect.collector_errors")
	metric.Value = 1
	if val, exists := entry.Data["collector"]; exists {
		metric.AddDimension("collector", val.(string))
	}

	writeToHandlers(hook.handlers, metric)
	return
}
