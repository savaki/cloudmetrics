package cloudmetrics

//	Copyright 2016 Matt Ho
//
//	Licensed under the Apache License, Version 2.0 (the "License");
//	you may not use this file except in compliance with the License.
//	You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
//	Unless required by applicable law or agreed to in writing, software
//	distributed under the License is distributed on an "AS IS" BASIS,
//	WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//	See the License for the specific language governing permissions and
//	limitations under the License.

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/rcrowley/go-metrics"
	"golang.org/x/net/context"
)

var debug = func(*Publisher) {}

var debug2 = Debug(os.Stderr)

type Mock struct {
	Inputs []*cloudwatch.PutMetricDataInput
}

func (m *Mock) PutMetricData(input *cloudwatch.PutMetricDataInput) (*cloudwatch.PutMetricDataOutput, error) {
	if m.Inputs == nil {
		m.Inputs = []*cloudwatch.PutMetricDataInput{}
	}
	m.Inputs = append(m.Inputs, input)
	return &cloudwatch.PutMetricDataOutput{}, nil
}

func TestPollOnceCounter(t *testing.T) {
	registry := metrics.NewRegistry()

	name := "my-metric"
	value := 5.0

	c := metrics.NewCounter()
	registry.Register(name, c)
	c.Inc(int64(value))

	publisher := newPublisher(registry, "blah", debug)
	data := publisher.pollOnce()

	if v := len(data); v != 1 {
		t.Errorf("expected 1 event to be published; got %v", v)
	}
	if v := data[0].MetricName; *v != name {
		t.Errorf("expected metricName to be %v; got %v", name, *v)
	}
	if v := data[0].Value; *v != value {
		t.Errorf("expected value to be %v; got %v", value, *v)
	}
}

func TestPollOnceGauge(t *testing.T) {
	registry := metrics.NewRegistry()

	value := 5.0

	c := metrics.NewGauge()
	registry.Register("blah", c)
	c.Update(int64(value))

	publisher := newPublisher(registry, "blah", debug)
	data := publisher.pollOnce()

	if v := len(data); v != 1 {
		t.Errorf("expected 1 event to be published; got %v", v)
	}
	if v := data[0].Value; *v != value {
		t.Errorf("expected value to be %v; got %v", value, *v)
	}
}

func TestPollOnceGauge64(t *testing.T) {
	registry := metrics.NewRegistry()

	value := 5.0

	c := metrics.NewGaugeFloat64()
	registry.Register("blah", c)
	c.Update(value)

	publisher := newPublisher(registry, "blah", debug)
	data := publisher.pollOnce()

	if v := len(data); v != 1 {
		t.Errorf("expected 1 event to be published; got %v", v)
	}
	if v := data[0].Value; *v != value {
		t.Errorf("expected value to be %v; got %v", value, *v)
	}
}

func TestPollOnceMeter(t *testing.T) {
	registry := metrics.NewRegistry()

	value := 5.0

	c := metrics.NewMeter()
	registry.Register("blah", c)
	c.Mark(int64(value))

	publisher := newPublisher(registry, "blah", debug)
	data := publisher.pollOnce()

	if v := len(data); v != 1 {
		t.Errorf("expected 1 event to be published; got %v", v)
	}
	if v := data[0].Value; *v != 0 {
		// for this number to be non-zero, this test would have to run for a while
		t.Errorf("expected Rate1 to be 0")
	}
}

func TestPollOnceHistogram(t *testing.T) {
	registry := metrics.NewRegistry()

	value := 2016.0
	c := metrics.NewHistogram(metrics.NewUniformSample(512))
	c.Update(int64(value))
	registry.Register("blah", c)

	publisher := newPublisher(registry, "blah", debug)
	data := publisher.pollOnce()
	sort.Sort(data)

	if v := len(data); v != 5 {
		t.Errorf("expected 1 event to be published; got %v", v)
	}

	if v := *data[0].MetricName; v != "blah.count" {
		t.Errorf("expected blah.count; got %v", v)
	}
	if v := *data[0].Value; v != 1 {
		t.Errorf("expected blah.count; got %v", v)
	}

	if v := *data[1].MetricName; v != "blah.p50" {
		t.Errorf("expected blah.p50; got %v", v)
	}
	if v := *data[1].Value; v != value {
		t.Errorf("expected blah.count; got %v", value)
	}

	if v := *data[2].MetricName; v != "blah.p75" {
		t.Errorf("expected blah.p75; got %v", v)
	}
	if v := *data[2].Value; v != value {
		t.Errorf("expected blah.count; got %v", value)
	}

	if v := *data[3].MetricName; v != "blah.p95" {
		t.Errorf("expected blah.p95; got %v", v)
	}
	if v := *data[3].Value; v != value {
		t.Errorf("expected blah.count; got %v", value)
	}

	if v := *data[4].MetricName; v != "blah.p99" {
		t.Errorf("expected blah.p99; got %v", v)
	}
	if v := *data[4].Value; v != value {
		t.Errorf("expected blah.count; got %v", value)
	}
}

func TestPollOnceHistogramCustomPercentile(t *testing.T) {
	registry := metrics.NewRegistry()

	value := 2016.0
	c := metrics.NewHistogram(metrics.NewUniformSample(512))
	c.Update(int64(value))
	registry.Register("blah", c)

	publisher := newPublisher(registry, "blah", debug, Percentiles([]float64{.44}))
	data := publisher.pollOnce()
	sort.Sort(data)

	if v := len(data); v != 2 {
		t.Errorf("expected 1 event to be published; got %v", v)
	}

	if v := *data[0].MetricName; v != "blah.count" {
		t.Errorf("expected blah.count; got %v", v)
	}
	if v := *data[0].Value; v != 1 {
		t.Errorf("expected blah.count; got %v", v)
	}

	if v := *data[1].MetricName; v != "blah.p44" {
		t.Errorf("expected blah.p44; got %v", v)
	}
	if v := *data[1].Value; v != value {
		t.Errorf("expected blah.count; got %v", value)
	}
}

func TestPollOnceTimer(t *testing.T) {
	registry := metrics.NewRegistry()

	value := float64(time.Millisecond * 200)
	c := metrics.NewTimer()
	c.Update(time.Duration(value))
	registry.Register("blah", c)

	publisher := newPublisher(registry, "blah", Interval(time.Millisecond*100), debug)
	data := publisher.pollOnce()
	sort.Sort(data)

	if v := len(data); v != 5 {
		t.Errorf("expected 1 event to be published; got %v", v)
	}

	if v := *data[0].MetricName; v != "blah.count" {
		t.Errorf("expected blah.count; got %v", v)
	}
	if v := *data[0].Value; v != 1 {
		t.Errorf("expected %v; got %v", value, v)
	}

	if v := *data[1].MetricName; v != "blah.p50" {
		t.Errorf("expected blah.p50; got %v", v)
	}
	if v := *data[1].Value; v != value {
		t.Errorf("expected %v; got %v", value, v)
	}

	if v := *data[2].MetricName; v != "blah.p75" {
		t.Errorf("expected blah.p75; got %v", v)
	}
	if v := *data[2].Value; v != value {
		t.Errorf("expected %v; got %v", value, v)
	}

	if v := *data[3].MetricName; v != "blah.p95" {
		t.Errorf("expected blah.p95; got %v", v)
	}
	if v := *data[3].Value; v != value {
		t.Errorf("expected %v; got %v", value, v)
	}

	if v := *data[4].MetricName; v != "blah.p99" {
		t.Errorf("expected blah.p99; got %v", v)
	}
	if v := *data[4].Value; v != value {
		t.Errorf("expected %v; got %v", value, v)
	}
}

func TestPollOnceDimensions(t *testing.T) {
	registry := metrics.NewRegistry()

	c := metrics.NewCounter()
	registry.Register("blah", c)
	c.Inc(1)

	publisher := newPublisher(registry, "blah", debug, Dimensions("foo", "bar", "baz", "qux"))
	data := publisher.pollOnce()

	if v := len(data[0].Dimensions); v != 2 {
		t.Errorf("expected 2 event dimensions; got %v", v)
	}

	d := data[0].Dimensions[0]
	if v := *d.Name; v != "foo" {
		t.Errorf("expected dimension name foo; got %v", v)
	}
	if v := *d.Value; v != "bar" {
		t.Errorf("expected dimension value bar; got %v", v)
	}

	d = data[0].Dimensions[1]
	if v := *d.Name; v != "baz" {
		t.Errorf("expected dimension name baz; got %v", v)
	}
	if v := *d.Value; v != "qux" {
		t.Errorf("expected dimension value qux; got %v", v)
	}
}

func TestPollOnceInvalidDimensions(t *testing.T) {
	registry := metrics.NewRegistry()

	c := metrics.NewCounter()
	registry.Register("blah", c)
	c.Inc(1)

	publisher := newPublisher(registry, "blah", debug, Dimensions("foo"))
	data := publisher.pollOnce()

	if v := len(data[0].Dimensions); v != 0 {
		t.Errorf("expected 0 dimensions; got %v", v)
	}
}

type NilWriter struct {
}

func (n NilWriter) Write(p []byte) (int, error) {
	return len(p), nil
}

func TestPublish(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(time.Millisecond * 200)
		cancel()
	}()

	namespace := "woot"

	registry := metrics.NewRegistry()

	// make many metrics
	for i := 0; i < 10; i++ {
		t := metrics.NewTimer()
		registry.Register(fmt.Sprintf("t%v", i), t)
		t.Update(time.Second)
	}

	mock := &Mock{}
	Publish(registry, namespace, Client(mock), Interval(time.Millisecond*50), Context(ctx), Debug(NilWriter{}))

	if len(mock.Inputs) == 0 {
		t.Error("expected at least one datum to have been published")
	}
}

func TestRegion(t *testing.T) {
	lookupFunc := func(context.Context) (io.ReadCloser, error) {
		return ioutil.NopCloser(strings.NewReader("us-east-1a")), nil
	}

	if v := region(lookupFunc); v != "us-east-1" {
		t.Errorf("expected us-east-1; got %v", v)
	}
}

func TestSortDatum(t *testing.T) {
	d1 := &cloudwatch.MetricDatum{MetricName: aws.String("a")}
	d2 := &cloudwatch.MetricDatum{MetricName: aws.String("b")}
	d3 := &cloudwatch.MetricDatum{MetricName: aws.String("c")}
	datums := Datums{d2, d1, d3}
	sort.Sort(datums)

	if v := datums[0]; v != d1 {
		t.Errorf("expected %v ; got %v", d1.MetricName, *v.MetricName)
	}
	if v := datums[1]; v != d2 {
		t.Errorf("expected %v ; got %v", d2.MetricName, *v.MetricName)
	}
	if v := datums[2]; v != d3 {
		t.Errorf("expected %v ; got %v", d3.MetricName, *v.MetricName)
	}
}
