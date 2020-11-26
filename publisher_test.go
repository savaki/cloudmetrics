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
	"errors"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"

	"github.com/aws/aws-sdk-go/service/cloudwatch"

	"github.com/gojuno/minimock/v3"
	"github.com/rcrowley/go-metrics"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/weareyolo/cloudmetrics/mock"
)

func TestPublisher__Publish(t *testing.T) {
	mc := minimock.NewController(t)
	defer mc.Finish()

	timer := metrics.NewTimer()
	timer.Update(1 * time.Second)

	registry := metrics.NewRegistry()
	err := registry.Register("timer", timer)
	require.NoError(t, err)

	mockedDatum := &cloudwatch.MetricDatum{
		MetricName: aws.String("timer"),
		Value:      aws.Float64(1),
		Unit:       aws.String(cloudwatch.StandardUnitCount),
	}

	b := mock.NewDatumBuilderMock(mc)
	b.BuildTimerDataMock.Expect(timer, "timer").Return([]*cloudwatch.MetricDatum{mockedDatum})

	t.Run("OK - No error from CW", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		logger, _ := test.NewNullLogger()

		cw := mock.NewCloudWatchMock(mc)
		cw.PutMetricDataMock.Expect(&cloudwatch.PutMetricDataInput{
			Namespace:  aws.String("nmsp"),
			MetricData: []*cloudwatch.MetricDatum{mockedDatum},
		}).Return(nil, nil)

		p := NewPublisher(registry, "nmsp",
			WithClient(cw),
			WithBuilder(b),
			WithInterval(6*time.Millisecond),
			WithContext(ctx),
			WithLogger(logger),
		)
		require.NotNil(t, p)

		go func() {
			time.Sleep(10 * time.Millisecond)
			cancel()
		}()

		p.Publish()

		require.NotZero(t, b.BuildTimerDataAfterCounter())
		require.NotZero(t, cw.PutMetricDataAfterCounter())

		assert.InDelta(t, 1, b.BuildTimerDataAfterCounter(), 1)
		assert.InDelta(t, 1, cw.PutMetricDataAfterCounter(), 1)
	})

	t.Run("OK - Error from CW is logged", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		logger, hook := test.NewNullLogger()

		cw := mock.NewCloudWatchMock(mc)
		cw.PutMetricDataMock.Expect(&cloudwatch.PutMetricDataInput{
			Namespace:  aws.String("nmsp"),
			MetricData: []*cloudwatch.MetricDatum{mockedDatum},
		}).Return(nil, errors.New("something happened"))

		p := NewPublisher(registry, "nmsp",
			WithClient(cw),
			WithBuilder(b),
			WithInterval(6*time.Millisecond),
			WithContext(ctx),
			WithLogger(logger),
		)
		require.NotNil(t, p)

		go func() {
			time.Sleep(10 * time.Millisecond)
			cancel()
		}()

		p.Publish()

		require.NotZero(t, b.BuildTimerDataAfterCounter())
		require.NotZero(t, cw.PutMetricDataAfterCounter())

		assert.InDelta(t, 1, b.BuildTimerDataAfterCounter(), 1)
		assert.InDelta(t, 1, cw.PutMetricDataAfterCounter(), 1)

		require.Len(t, hook.Entries, 1)
		entry := hook.Entries[0]
		assert.Equal(t, logrus.ErrorLevel, entry.Level)
		assert.Equal(t, "could not put last chunk of metrics", entry.Message)
		assert.EqualError(t, entry.Data[logrus.ErrorKey].(error), "something happened")
	})
}
