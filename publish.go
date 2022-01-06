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
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/rcrowley/go-metrics"
	"golang.org/x/net/context"
	"golang.org/x/net/context/ctxhttp"
)

// CloudWatch is an interface for *cloudwatch.CloudWatch that clearly identifies the functions
// used by cloudmetrics
type CloudWatch interface {
	PutMetricData(input *cloudwatch.PutMetricDataInput) (*cloudwatch.PutMetricDataOutput, error)
}

type Datums []*cloudwatch.MetricDatum

func (d Datums) Len() int {
	return len(d)
}

func (d Datums) Less(i, j int) bool {
	return *d[i].MetricName < *d[j].MetricName
}

func (d Datums) Swap(i, j int) {
	t := d[i]
	d[i] = d[j]
	d[j] = t
}

func lookupAvailabilityZone(ctx context.Context) (io.ReadCloser, error) {
	resp, err := ctxhttp.Get(ctx, http.DefaultClient, "http://169.254.169.254/latest/meta-data/placement/availability-zone")
	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}

type Publisher struct {
	ctx         context.Context
	registry    metrics.Registry
	client      CloudWatch
	interval    time.Duration
	percentiles []float64
	namespace   *string
	debug       func(v ...interface{})
	dimensions  []*cloudwatch.Dimension
	ch          chan *cloudwatch.MetricDatum
}

func (p *Publisher) putMetrics(data []*cloudwatch.MetricDatum) error {
	_, err := p.client.PutMetricData(&cloudwatch.PutMetricDataInput{
		Namespace:  p.namespace,
		MetricData: data,
	})
	return err
}

func (p *Publisher) publishMetrics(data []*cloudwatch.MetricDatum) {
	for len(data) > 20 {
		p.putMetrics(data[0:20])
		data = data[20:]
	}

	if len(data) > 0 {
		p.putMetrics(data)
	}
}

func (p *Publisher) pollOnce() Datums {
	p.debug("Polling metrics")

	now := aws.Time(time.Now())
	data := []*cloudwatch.MetricDatum{}

	build := func(name string) *cloudwatch.MetricDatum {
		p.debug("Building metric,", name)
		return &cloudwatch.MetricDatum{
			MetricName: aws.String(name),
			Dimensions: p.dimensions,
			Timestamp:  now,
		}
	}

	p.registry.Each(func(name string, i interface{}) {
		switch v := i.(type) {

		case metrics.Counter:
			count := float64(v.Count())
			datum := build(name)
			datum.Unit = aws.String(cloudwatch.StandardUnitCount)
			datum.Value = aws.Float64(count)
			data = append(data, datum)

		case metrics.Gauge:
			value := float64(v.Value())
			datum := build(name)
			datum.Unit = aws.String(cloudwatch.StandardUnitCount)
			datum.Value = aws.Float64(value)
			data = append(data, datum)

		case metrics.GaugeFloat64:
			value := float64(v.Value())
			datum := build(name)
			datum.Unit = aws.String(cloudwatch.StandardUnitCount)
			datum.Value = aws.Float64(value)
			data = append(data, datum)

		case metrics.Histogram:
			metric := v.Snapshot()
			if metric.Count() == 0 {
				return
			}
			points := map[string]float64{
				fmt.Sprintf("%s.count", name): float64(metric.Count()),
			}
			for index, pct := range metric.Percentiles(p.percentiles) {
				k := fmt.Sprintf("%s.p%v", name, int(p.percentiles[index]*100))
				points[k] = pct
			}
			for n, v := range points {
				datum := build(n)
				datum.Value = aws.Float64(v)
				data = append(data, datum)
			}

		case metrics.Meter:
			value := v.Rate1()
			datum := build(name)
			datum.Unit = aws.String(cloudwatch.StandardUnitCount)
			datum.Value = aws.Float64(value)
			data = append(data, datum)

		case metrics.Timer:
			metric := v.Snapshot()
			if metric.Count() == 0 {
				return
			}
			points := map[string]float64{
				fmt.Sprintf("%s.count", name): float64(metric.Count()),
			}
			percentiles := []float64{.5, .75, .95, .99}
			for index, pct := range metric.Percentiles(percentiles) {
				k := fmt.Sprintf("%s.p%v", name, int(percentiles[index]*100))
				points[k] = pct
			}
			for n, v := range points {
				datum := build(n)
				datum.Value = aws.Float64(v)
				data = append(data, datum)
			}

		default:
			p.debug(fmt.Sprintf("Received unexpected metric, %#v", i))
			return
		}
	})

	p.debug("Received", len(data), "event(s)")

	return data
}

func (p *Publisher) pollMetrics() {
	ticker := time.NewTicker(p.interval)
	defer ticker.Stop()

	for {
		p.debug("Waiting", p.interval)
		// 1. Wait for either a tick or the context to close
		select {
		case <-p.ctx.Done():
			return
		case <-ticker.C:
		}

		data := p.pollOnce()
		p.publishMetrics(data)
	}
}

func region(lookupAZ func(context.Context) (io.ReadCloser, error)) string {
	region := os.Getenv("AWS_REGION")

	if region == "" {
		region = os.Getenv("AWS_DEFAULT_REGION")
	}

	if region == "" {
		ctx, _ := context.WithTimeout(context.Background(), time.Second)
		body, err := lookupAZ(ctx)
		if err == nil {
			defer body.Close()

			data, err := ioutil.ReadAll(body)
			if err == nil {
				region = strings.TrimSpace(string(data))
				if len(region) > 0 {
					region = region[0 : len(region)-1]
				}
			}
		}
	}

	if region == "" {
		region = "us-east-1"
	}

	return region
}

func client() CloudWatch {
	cfg := &aws.Config{Region: aws.String(region(lookupAvailabilityZone))}

	return cloudwatch.New(session.New(cfg))
}

func newPublisher(registry metrics.Registry, namespace string, configs ...func(*Publisher)) *Publisher {
	publisher := Publisher{
		ctx:         context.Background(),
		registry:    registry,
		namespace:   aws.String(namespace),
		client:      client(),
		interval:    time.Minute,
		percentiles: []float64{.5, .75, .95, .99},
		debug:       func(...interface{}) {},
		dimensions:  []*cloudwatch.Dimension{},
		ch:          make(chan *cloudwatch.MetricDatum, 4096),
	}

	for _, config := range configs {
		config(&publisher)
	}

	return &publisher
}

// Publish is the main entry point to publish metrics on a recurring basis to CloudWatch
func Publish(registry metrics.Registry, namespace string, configs ...func(*Publisher)) {
	publisher := newPublisher(registry, namespace, configs...)
	publisher.pollMetrics()
}

// Interval allows for a custom posting interval; by default, the interval is every 1 minute
func Interval(interval time.Duration) func(*Publisher) {
	return func(p *Publisher) {
		p.interval = interval
	}
}

// Dimensions allows for user specified dimensiosn to be added to the post
func Dimensions(keyvals ...string) func(*Publisher) {
	return func(p *Publisher) {
		if len(keyvals)%2 != 0 {
			fmt.Fprintf(os.Stderr, "Dimensions requires an even number of arguments")
			return
		}

		for i := 0; i < len(keyvals); i = i + 2 {
			p.dimensions = append(p.dimensions, &cloudwatch.Dimension{
				Name:  aws.String(keyvals[i]),
				Value: aws.String(keyvals[i+1]),
			})
		}
	}
}

// Percentiles allows the reported percentiles for Histogram and Timer metrics to be customized
func Percentiles(percentiles []float64) func(*Publisher) {
	return func(p *Publisher) {
		p.percentiles = percentiles
	}
}

// Client allows for user provided *cloudwatch.Cloudwatch instance
func Client(client CloudWatch) func(*Publisher) {
	return func(p *Publisher) {
		p.client = client
	}
}

// Context allows a context to be specified.  When <-ctx.Done() returns; the Publisher will
// stop any internal go routines and return
func Context(ctx context.Context) func(*Publisher) {
	return func(p *Publisher) {
		p.ctx = ctx
	}
}

// Debug writes additional data to the writer specified
func Debug(w io.Writer) func(*Publisher) {
	return func(p *Publisher) {
		p.debug = func(args ...interface{}) {
			fmt.Fprintln(w, args...)
		}
	}
}
