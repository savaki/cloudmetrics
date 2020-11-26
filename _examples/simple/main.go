package main

import (
	"os"
	"time"

	"github.com/rcrowley/go-metrics"
	"github.com/weareyolo/cloudmetrics"
)

func main() {
	t := metrics.NewTimer()
	metrics.Register("sample", t)
	t.Update(time.Millisecond)

	// publish metrics to cloudwatch
	cloudmetrics.Publish(metrics.DefaultRegistry, "sample-namespace",
		cloudmetrics.Interval(5*time.Second),
		cloudmetrics.Debug(os.Stderr),
	)
}
