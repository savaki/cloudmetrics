# cloudmetrics

[![GoDoc](https://godoc.org/github.com/savaki/cloudmetrics?status.svg)](https://godoc.org/github.com/savaki/cloudmetrics)

This is a reporter for the [go-metrics](https://github.com/rcrowley/go-metrics)
that will posts metrics to [CloudWatch](https://aws.amazon.com/cloudwatch/).

## Usage

```go
import "github.com/weareyolo/cloudmetrics"

go cloudmetrics.Publish(metrics.DefaultRegistry,
    "/sample/", // namespace
)
```

## Configuration

`cloudmetrics` supports a number of configuration options:

```go
package main

import (
    "context"
    "time"

    "github.com/aws/aws-sdk-go/service/cloudwatch"
    "github.com/rcrowley/go-metrics"
    "github.com/weareyolo/cloudmetrics"
)

func main() {
    p := cloudmetrics.NewPublisher(
        metrics.DefaultRegistry,                            // metrics registry
        "/sample/",                                         // namespace
        cloudmetrics.WithDimensions(map[string]string{
            "k1": "v1",
            "k2": "v2",
        }),                                                 // allows for custom dimensions
        cloudmetrics.WithInterval(5*time.Minute),           // custom interval
        cloudmetrics.WithContext(context.Background()),     // enables graceful shutdown via contexts
        cloudmetrics.WithPercentiles([]float64{.5, .99}),   // customize percentiles for histograms and timers
        cloudmetrics.WithUnits(map[string]string{
            "size": cloudwatch.StandardUnitGigabytes,
        }),                                                 // customize units based on metric names
    )
    go p.Publish()
    for {
        time.Sleep(5 * time.Minute)
    }
}

```
