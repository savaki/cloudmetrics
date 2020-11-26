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
import "github.com/weareyolo/cloudmetrics"

go cloudmetrics.Publish(metrics.DefaultRegistry,
    "/sample/",                                      // namespace
    cloudmetrics.Dimensions("k1", "v1", "k2", "v2"), // allows for custom dimensions
    cloudmetrics.Interval(time.Minutes * 5),         // custom interval
    cloudmetrics.Context(context.Background()),      // enables graceful shutdown via contexts
    cloudmetrics.Percentiles([]float64{.5, .99}),    // customize percentiles for histograms and timers
)

```
