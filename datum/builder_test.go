package datum

import (
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/rcrowley/go-metrics"
	"github.com/stretchr/testify/assert"
)

func TestBuilder__BuildCounterData(t *testing.T) {
	name := "my-metric"
	value := 5.0
	m := metrics.NewCounter()
	m.Inc(int64(value))

	t.Run("OK - No config", func(t *testing.T) {
		b := NewBuilder(nil, nil, nil, 30)
		data := b.BuildCounterData(m, name)

		assert.Len(t, data, 1)
		assert.Equal(t, &cloudwatch.MetricDatum{
			MetricName:        &name,
			Value:             &value,
			Unit:              aws.String(cloudwatch.StandardUnitCount),
			Timestamp:         data[0].Timestamp,
			StorageResolution: aws.Int64(30),
		}, data[0])
	})

	t.Run("OK - With dimensions", func(t *testing.T) {
		b := NewBuilder(nil, map[string]string{"foo": "bar"}, nil, 30)
		data := b.BuildCounterData(m, name)

		assert.Len(t, data, 1)
		assert.Equal(t, &cloudwatch.MetricDatum{
			MetricName: &name,
			Value:      &value,
			Unit:       aws.String(cloudwatch.StandardUnitCount),
			Dimensions: []*cloudwatch.Dimension{
				{
					Name:  aws.String("foo"),
					Value: aws.String("bar"),
				},
			},
			Timestamp:         data[0].Timestamp,
			StorageResolution: aws.Int64(30),
		}, data[0])
	})
}

func TestBuilder__BuildGaugeData(t *testing.T) {
	name := "my-metric"
	value := 5.0
	m := metrics.NewGauge()
	m.Update(int64(value))

	t.Run("OK - No config", func(t *testing.T) {
		b := NewBuilder(nil, nil, nil, 30)
		data := b.BuildGaugeData(m, name)

		assert.Len(t, data, 1)
		assert.Equal(t, &cloudwatch.MetricDatum{
			MetricName:        &name,
			Value:             &value,
			Unit:              aws.String(cloudwatch.StandardUnitCount),
			Timestamp:         data[0].Timestamp,
			StorageResolution: aws.Int64(30),
		}, data[0])
	})
}

func TestBuilder__BuildGaugeFloat64Data(t *testing.T) {
	name := "my-metric"
	value := 5.0
	m := metrics.NewGaugeFloat64()
	m.Update(value)

	t.Run("OK - No config", func(t *testing.T) {
		b := NewBuilder(nil, nil, nil, 30)
		data := b.BuildGaugeFloat64Data(m, name)

		assert.Len(t, data, 1)
		assert.Equal(t, &cloudwatch.MetricDatum{
			MetricName:        &name,
			Value:             &value,
			Unit:              aws.String(cloudwatch.StandardUnitCount),
			Timestamp:         data[0].Timestamp,
			StorageResolution: aws.Int64(30),
		}, data[0])
	})
}

func TestBuilder__BuildMeterData(t *testing.T) {
	name := "my-metric"
	value := 5.0
	m := metrics.NewMeter()
	m.Mark(int64(value))

	t.Run("OK - No config", func(t *testing.T) {
		b := NewBuilder(nil, nil, nil, 30)
		data := b.BuildMeterData(m, name)

		assert.Len(t, data, 1)
		assert.Equal(t, &cloudwatch.MetricDatum{
			MetricName:        &name,
			Value:             aws.Float64(0),
			Unit:              aws.String(cloudwatch.StandardUnitCount),
			Timestamp:         data[0].Timestamp,
			StorageResolution: aws.Int64(30),
		}, data[0])
	})
}

func TestBuilder__BuildHistogramData(t *testing.T) {
	name := "my-metric"
	value := 2016.0
	m := metrics.NewHistogram(metrics.NewUniformSample(512))
	m.Update(int64(value))

	t.Run("OK - No config", func(t *testing.T) {
		b := NewBuilder(nil, nil, nil, 30)
		data := b.BuildHistogramData(m, name)

		assert.Len(t, data, 1)
		assert.Equal(t, &cloudwatch.MetricDatum{
			MetricName:        aws.String(name + ".count"),
			Value:             aws.Float64(1),
			Unit:              aws.String(cloudwatch.StandardUnitCount),
			Timestamp:         data[0].Timestamp,
			StorageResolution: aws.Int64(30),
		}, data[0])
	})

	t.Run("OK - With percentiles", func(t *testing.T) {
		b := NewBuilder(nil, nil, []float64{0.44}, 30)
		data := b.BuildHistogramData(m, name)

		assert.Len(t, data, 2)
		tmstp := data[0].Timestamp
		assert.ElementsMatch(t, []*cloudwatch.MetricDatum{
			{
				MetricName:        aws.String(name + ".count"),
				Value:             aws.Float64(1),
				Unit:              aws.String(cloudwatch.StandardUnitCount),
				Timestamp:         tmstp,
				StorageResolution: aws.Int64(30),
			},
			{
				MetricName:        aws.String(name + ".p44"),
				Value:             &value,
				Unit:              aws.String(cloudwatch.StandardUnitCount),
				Timestamp:         tmstp,
				StorageResolution: aws.Int64(30),
			},
		}, data)
	})
}

func TestBuilder__BuildTimerData(t *testing.T) {
	name := "my-metric"
	m := metrics.NewTimer()
	m.Update(time.Duration(200) * time.Millisecond)

	t.Run("OK - No config", func(t *testing.T) {
		b := NewBuilder(nil, nil, nil, 30)
		data := b.BuildTimerData(m, name)

		assert.Len(t, data, 1)
		assert.Equal(t, &cloudwatch.MetricDatum{

			MetricName:        aws.String(name + ".count"),
			Value:             aws.Float64(1),
			Unit:              aws.String(cloudwatch.StandardUnitCount),
			Timestamp:         data[0].Timestamp,
			StorageResolution: aws.Int64(30),
		}, data[0])
	})

	t.Run("OK - With percentiles, default unit is ms", func(t *testing.T) {
		b := NewBuilder(nil, nil, []float64{0.5}, 30)
		data := b.BuildTimerData(m, name)

		assert.Len(t, data, 2)
		tmstp := data[0].Timestamp
		assert.ElementsMatch(t, []*cloudwatch.MetricDatum{
			{
				MetricName:        aws.String(name + ".count"),
				Value:             aws.Float64(1),
				Unit:              aws.String(cloudwatch.StandardUnitCount),
				Timestamp:         tmstp,
				StorageResolution: aws.Int64(30),
			},
			{
				MetricName:        aws.String(name + ".p50"),
				Value:             aws.Float64(200),
				Unit:              aws.String(cloudwatch.StandardUnitMilliseconds),
				Timestamp:         tmstp,
				StorageResolution: aws.Int64(30),
			},
		}, data)
	})

	t.Run("OK - With percentiles and None unit, time.Duration = ns", func(t *testing.T) {
		b := NewBuilder(
			map[string]string{
				name: cloudwatch.StandardUnitNone,
			},
			nil,
			[]float64{0.5},
			30,
		)
		data := b.BuildTimerData(m, name)

		assert.Len(t, data, 2)
		tmstp := data[0].Timestamp
		assert.ElementsMatch(t, []*cloudwatch.MetricDatum{
			{
				MetricName:        aws.String(name + ".count"),
				Value:             aws.Float64(1),
				Unit:              aws.String(cloudwatch.StandardUnitCount),
				Timestamp:         tmstp,
				StorageResolution: aws.Int64(30),
			},
			{
				MetricName:        aws.String(name + ".p50"),
				Value:             aws.Float64(2e8),
				Unit:              aws.String(cloudwatch.StandardUnitNone),
				Timestamp:         tmstp,
				StorageResolution: aws.Int64(30),
			},
		}, data)
	})

	t.Run("OK - All config, second unit", func(t *testing.T) {
		dims := []*cloudwatch.Dimension{
			{
				Name:  aws.String("dim"),
				Value: aws.String("val"),
			},
		}
		b := NewBuilder(
			map[string]string{
				name: cloudwatch.StandardUnitSeconds,
			},
			map[string]string{"dim": "val"},
			[]float64{0.5},
			30,
		)
		data := b.BuildTimerData(m, name)

		assert.Len(t, data, 2)
		tmstp := data[0].Timestamp

		assert.ElementsMatch(t, []*cloudwatch.MetricDatum{
			{
				MetricName:        aws.String(name + ".count"),
				Value:             aws.Float64(1),
				Unit:              aws.String(cloudwatch.StandardUnitCount),
				Dimensions:        dims,
				Timestamp:         tmstp,
				StorageResolution: aws.Int64(30),
			},
			{
				MetricName:        aws.String(name + ".p50"),
				Value:             aws.Float64(0.2),
				Unit:              aws.String(cloudwatch.StandardUnitSeconds),
				Dimensions:        dims,
				Timestamp:         tmstp,
				StorageResolution: aws.Int64(30),
			},
		}, data)
	})
}
