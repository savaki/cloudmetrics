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
	"context"
	"time"

	awscloudmetrics "github.com/weareyolo/cloudmetrics/aws"
	"github.com/weareyolo/cloudmetrics/datum"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/rcrowley/go-metrics"
	"github.com/sirupsen/logrus"
)

// AWS limitation on `MetricData` length in `PutMetricDataInput`
const batchSize = 20

type publisher struct {
	ctx          context.Context
	registry     metrics.Registry
	client       CloudWatch
	interval     time.Duration
	namespace    *string
	logger       logrus.FieldLogger
	datumBuilder DatumBuilder
}

// NewPublisher creates a configured Publisher
func NewPublisher(registry metrics.Registry, namespace string, opts ...Option) Publisher {
	s := getSettings(opts)

	b := s.DatumBuilder
	if b == nil {
		b = datum.NewBuilder(s.Units, s.Dimensions, s.Percentiles, s.StorageResolution)
	}

	c := s.Client
	if c == nil {
		c = awscloudmetrics.NewCloudWatchClient()
	}

	l := s.Logger
	if l == nil {
		l = newLogger()
	}

	return &publisher{
		ctx:          s.Context,
		registry:     registry,
		namespace:    aws.String(namespace),
		client:       c,
		interval:     s.Interval,
		logger:       l,
		datumBuilder: b,
	}
}

// Publish is the main entry point to publish metrics on a recurring basis to CloudWatch.
func (p *publisher) Publish() {
	ticker := time.NewTicker(p.interval)
	defer ticker.Stop()

	for {
		p.logger.Debugf("Waiting %v", p.interval)
		// 1. Wait for either a tick or the context to close
		select {
		case <-p.ctx.Done():
			return
		case <-ticker.C:
		}

		p.publishMetrics(p.pollOnce())
	}
}

func (p *publisher) pollOnce() []*cloudwatch.MetricDatum {
	p.logger.Debug("Polling metrics")
	data := []*cloudwatch.MetricDatum{}

	p.registry.Each(func(name string, i interface{}) {
		switch v := i.(type) {

		case metrics.Counter:
			data = append(data, p.datumBuilder.BuildCounterData(v, name)...)

		case metrics.Gauge:
			data = append(data, p.datumBuilder.BuildGaugeData(v, name)...)

		case metrics.GaugeFloat64:
			data = append(data, p.datumBuilder.BuildGaugeFloat64Data(v, name)...)

		case metrics.Histogram:
			data = append(data, p.datumBuilder.BuildHistogramData(v, name)...)

		case metrics.Meter:
			data = append(data, p.datumBuilder.BuildMeterData(v, name)...)

		case metrics.Timer:
			data = append(data, p.datumBuilder.BuildTimerData(v, name)...)

		default:
			p.logger.Errorf("Received unexpected metric: %#v", i)
			return
		}
	})

	p.logger.Debugf("Received %v event(s)", len(data))
	return data
}

func (p *publisher) publishMetrics(data []*cloudwatch.MetricDatum) {
	for len(data) > batchSize {
		if err := p.putMetrics(data[0:batchSize]); err != nil {
			p.logger.WithError(err).Error("could not put chunk of metrics")
		}
		data = data[batchSize:]
	}

	if len(data) > 0 {
		if err := p.putMetrics(data); err != nil {
			p.logger.WithError(err).Error("could not put last chunk of metrics")
		}
	}
}

func (p *publisher) putMetrics(data []*cloudwatch.MetricDatum) error {
	_, err := p.client.PutMetricData(&cloudwatch.PutMetricDataInput{
		Namespace:  p.namespace,
		MetricData: data,
	})
	return err
}
