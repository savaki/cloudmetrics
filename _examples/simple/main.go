package main

import (
	"time"

	"github.com/sirupsen/logrus"

	"github.com/rcrowley/go-metrics"
	"github.com/weareyolo/cloudmetrics"
)

func main() {
	t := metrics.NewTimer()
	metrics.Register("sample", t)
	t.Update(time.Millisecond)

	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	// publish metrics to cloudwatch
	p := cloudmetrics.NewPublisher(metrics.DefaultRegistry, "sample-namespace",
		cloudmetrics.WithInterval(5*time.Second),
		cloudmetrics.WithLogger(logger),
	)
	p.Publish()
}
