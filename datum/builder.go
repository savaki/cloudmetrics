package datum

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/rcrowley/go-metrics"
)

var convertDuration = map[string](func(v float64) float64){
	cloudwatch.StandardUnitSeconds:      func(v float64) float64 { return v / float64(time.Second) },
	cloudwatch.StandardUnitMilliseconds: func(v float64) float64 { return v / float64(time.Millisecond) },
	cloudwatch.StandardUnitMicroseconds: func(v float64) float64 { return v / float64(time.Microsecond) },
}

func identity(v float64) float64 { return v }

// Builder handles the datum generation
type Builder struct {
	units             map[string]string
	dimensions        []*cloudwatch.Dimension
	percentiles       []float64
	storageResolution *int64
}

// NewBuilder creates a Builder
func NewBuilder(units map[string]string, dimensions map[string]string, percentiles []float64,
	storageResolution int64) *Builder {

	var dims []*cloudwatch.Dimension = nil
	n := len(dimensions)
	if n > 0 {
		dims = make([]*cloudwatch.Dimension, 0, len(dimensions))
		for k, v := range dimensions {
			dims = append(dims, &cloudwatch.Dimension{
				Name:  aws.String(k),
				Value: aws.String(v),
			})
		}
	}

	return &Builder{
		units:             units,
		dimensions:        dims,
		percentiles:       percentiles,
		storageResolution: aws.Int64(storageResolution),
	}
}

func (b *Builder) buildDatum(name string, value float64, unit string, t time.Time) *cloudwatch.MetricDatum {
	return &cloudwatch.MetricDatum{
		MetricName:        aws.String(name),
		Value:             aws.Float64(value),
		Unit:              aws.String(unit),
		Dimensions:        b.dimensions,
		Timestamp:         aws.Time(t.UTC()),
		StorageResolution: b.storageResolution,
	}
}

func (b *Builder) getMetricUnit(name string, defaultUnit string) string {
	if val, ok := b.units[name]; ok {
		return val
	}
	return defaultUnit
}

// BuildCounterData generates data from a Counter
func (b *Builder) BuildCounterData(v metrics.Counter, name string) []*cloudwatch.MetricDatum {
	unit := b.getMetricUnit(name, cloudwatch.StandardUnitCount)
	datum := b.buildDatum(name, float64(v.Count()), unit, time.Now())
	return []*cloudwatch.MetricDatum{datum}
}

// BuildGaugeData generates data from a Gauge
func (b *Builder) BuildGaugeData(v metrics.Gauge, name string) []*cloudwatch.MetricDatum {
	unit := b.getMetricUnit(name, cloudwatch.StandardUnitCount)
	datum := b.buildDatum(name, float64(v.Value()), unit, time.Now())
	return []*cloudwatch.MetricDatum{datum}
}

// BuildGaugeFloat64Data generates data from a GaugeFloat64
func (b *Builder) BuildGaugeFloat64Data(v metrics.GaugeFloat64, name string) []*cloudwatch.MetricDatum {
	unit := b.getMetricUnit(name, cloudwatch.StandardUnitCount)
	datum := b.buildDatum(name, float64(v.Value()), unit, time.Now())
	return []*cloudwatch.MetricDatum{datum}
}

// BuildMeterData generates data from a Meter
func (b *Builder) BuildMeterData(v metrics.Meter, name string) []*cloudwatch.MetricDatum {
	unit := b.getMetricUnit(name, cloudwatch.StandardUnitCount)
	datum := b.buildDatum(name, float64(v.Rate1()), unit, time.Now())
	return []*cloudwatch.MetricDatum{datum}
}

// BuildHistogramData generates data from an Histogram
func (b *Builder) BuildHistogramData(v metrics.Histogram, name string) []*cloudwatch.MetricDatum {
	metric := v.Snapshot()
	if metric.Count() == 0 {
		return nil
	}

	t := time.Now()
	// Build Count datum
	datum := b.buildDatum(fmt.Sprintf("%s.count", name), float64(metric.Count()), cloudwatch.StandardUnitCount, t)
	res := []*cloudwatch.MetricDatum{datum}

	// Build Percentiles data
	unit := b.getMetricUnit(name, cloudwatch.StandardUnitCount)
	for index, val := range metric.Percentiles(b.percentiles) {
		n := fmt.Sprintf("%s.p%v", name, int(b.percentiles[index]*100))
		datum := b.buildDatum(n, val, unit, t)
		res = append(res, datum)
	}

	return res
}

// BuildTimerData generates data from a Timer
func (b *Builder) BuildTimerData(v metrics.Timer, name string) []*cloudwatch.MetricDatum {
	metric := v.Snapshot()
	if metric.Count() == 0 {
		return nil
	}

	t := time.Now()
	// Build Count datum
	datum := b.buildDatum(fmt.Sprintf("%s.count", name), float64(metric.Count()), cloudwatch.StandardUnitCount, t)
	res := []*cloudwatch.MetricDatum{datum}

	// Build Percentiles data
	unit := b.getMetricUnit(name, cloudwatch.StandardUnitMilliseconds)
	convertFunc := identity
	if cf, ok := convertDuration[unit]; ok {
		convertFunc = cf
	}

	for index, val := range metric.Percentiles(b.percentiles) {
		n := fmt.Sprintf("%s.p%v", name, int(b.percentiles[index]*100))
		datum := b.buildDatum(n, convertFunc(val), unit, t)
		res = append(res, datum)
	}

	return res
}
