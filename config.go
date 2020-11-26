package cloudmetrics

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"
)

type settings struct {
	Context           context.Context
	Client            CloudWatch
	Interval          time.Duration
	Logger            logrus.FieldLogger
	Units             map[string]string
	Dimensions        map[string]string
	Percentiles       []float64
	StorageResolution int64
	DatumBuilder      DatumBuilder
}

// Option is a type made to override default values for Publisher
type Option func(s *settings)

// WithContext allows a context to be specified.  When <-ctx.Done() returns; the Publisher will
// stop any internal go routines and return
func WithContext(ctx context.Context) Option {
	return func(s *settings) {
		s.Context = ctx
	}
}

// WithClient allows for user provided *cloudwatch.Cloudwatch instance
func WithClient(client CloudWatch) Option {
	return func(s *settings) {
		s.Client = client
	}
}

// WithInterval allows for a custom posting interval; by default, the interval is every 1 minute
func WithInterval(interval time.Duration) Option {
	return func(s *settings) {
		s.Interval = interval
	}
}

// WithLogger allows to use custom logger
func WithLogger(logger logrus.FieldLogger) Option {
	return func(s *settings) {
		s.Logger = logger
	}
}

// WithUnits specifies the AWS StandardUnits to use for given metrics
func WithUnits(units map[string]string) Option {
	return func(s *settings) {
		for k, v := range units {
			s.Units[k] = v
		}
	}
}

// WithDimensions allows for user specified dimensions to be added to the post
func WithDimensions(dimensions map[string]string) Option {
	return func(s *settings) {
		s.Dimensions = dimensions
	}
}

// WithPercentiles allows the reported percentiles for Histogram and Timer metrics to be customized
func WithPercentiles(percentiles []float64) Option {
	return func(s *settings) {
		s.Percentiles = percentiles
	}
}

// WithStorageResolution specifies the Storage Resolution to use in seconds, default to 60
func WithStorageResolution(storageResolution int64) Option {
	return func(s *settings) {
		s.StorageResolution = storageResolution
	}
}

// WithBuilder specifies the DatumBuilder to use
func WithBuilder(b DatumBuilder) Option {
	return func(s *settings) {
		s.DatumBuilder = b
	}
}

func getSettings(opts []Option) *settings {
	s := &settings{
		Context:           context.Background(),
		Interval:          1 * time.Minute,
		Units:             map[string]string{},
		Dimensions:        map[string]string{},
		Percentiles:       []float64{.5, .75, .95, .99},
		StorageResolution: 60,
	}

	for _, o := range opts {
		o(s)
	}

	return s
}

func newLogger() logrus.FieldLogger {
	l := logrus.New()
	l.SetReportCaller(true)
	l.SetFormatter(new(logrus.JSONFormatter))
	l.SetLevel(logrus.ErrorLevel)
	return l
}
