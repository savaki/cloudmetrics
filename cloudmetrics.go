package cloudmetrics

import (
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/rcrowley/go-metrics"
)

// Publisher handles the publication of metrics data to CloudWatch
type Publisher interface {
	Publish()
}

// DatumBuilder handles the datum generation per metric type
type DatumBuilder interface {
	BuildCounterData(v metrics.Counter, name string) []*cloudwatch.MetricDatum
	BuildGaugeData(v metrics.Gauge, name string) []*cloudwatch.MetricDatum
	BuildGaugeFloat64Data(v metrics.GaugeFloat64, name string) []*cloudwatch.MetricDatum
	BuildMeterData(v metrics.Meter, name string) []*cloudwatch.MetricDatum
	BuildHistogramData(v metrics.Histogram, name string) []*cloudwatch.MetricDatum
	BuildTimerData(v metrics.Timer, name string) []*cloudwatch.MetricDatum
}

// CloudWatch is an interface for *cloudwatch.CloudWatch that clearly identifies the functions
// used by cloudmetrics
type CloudWatch interface {
	PutMetricData(input *cloudwatch.PutMetricDataInput) (*cloudwatch.PutMetricDataOutput, error)
}
