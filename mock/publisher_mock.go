package mock

// Code generated by http://github.com/gojuno/minimock (3.0.6). DO NOT EDIT.

import (
	mm_atomic "sync/atomic"
	mm_time "time"

	"github.com/gojuno/minimock/v3"
)

// PublisherMock implements cloudmetrics.Publisher
type PublisherMock struct {
	t minimock.Tester

	funcPublish          func()
	inspectFuncPublish   func()
	afterPublishCounter  uint64
	beforePublishCounter uint64
	PublishMock          mPublisherMockPublish
}

// NewPublisherMock returns a mock for cloudmetrics.Publisher
func NewPublisherMock(t minimock.Tester) *PublisherMock {
	m := &PublisherMock{t: t}
	if controller, ok := t.(minimock.MockController); ok {
		controller.RegisterMocker(m)
	}

	m.PublishMock = mPublisherMockPublish{mock: m}

	return m
}

type mPublisherMockPublish struct {
	mock               *PublisherMock
	defaultExpectation *PublisherMockPublishExpectation
	expectations       []*PublisherMockPublishExpectation
}

// PublisherMockPublishExpectation specifies expectation struct of the Publisher.Publish
type PublisherMockPublishExpectation struct {
	mock *PublisherMock

	Counter uint64
}

// Expect sets up expected params for Publisher.Publish
func (mmPublish *mPublisherMockPublish) Expect() *mPublisherMockPublish {
	if mmPublish.mock.funcPublish != nil {
		mmPublish.mock.t.Fatalf("PublisherMock.Publish mock is already set by Set")
	}

	if mmPublish.defaultExpectation == nil {
		mmPublish.defaultExpectation = &PublisherMockPublishExpectation{}
	}

	return mmPublish
}

// Inspect accepts an inspector function that has same arguments as the Publisher.Publish
func (mmPublish *mPublisherMockPublish) Inspect(f func()) *mPublisherMockPublish {
	if mmPublish.mock.inspectFuncPublish != nil {
		mmPublish.mock.t.Fatalf("Inspect function is already set for PublisherMock.Publish")
	}

	mmPublish.mock.inspectFuncPublish = f

	return mmPublish
}

// Return sets up results that will be returned by Publisher.Publish
func (mmPublish *mPublisherMockPublish) Return() *PublisherMock {
	if mmPublish.mock.funcPublish != nil {
		mmPublish.mock.t.Fatalf("PublisherMock.Publish mock is already set by Set")
	}

	if mmPublish.defaultExpectation == nil {
		mmPublish.defaultExpectation = &PublisherMockPublishExpectation{mock: mmPublish.mock}
	}

	return mmPublish.mock
}

//Set uses given function f to mock the Publisher.Publish method
func (mmPublish *mPublisherMockPublish) Set(f func()) *PublisherMock {
	if mmPublish.defaultExpectation != nil {
		mmPublish.mock.t.Fatalf("Default expectation is already set for the Publisher.Publish method")
	}

	if len(mmPublish.expectations) > 0 {
		mmPublish.mock.t.Fatalf("Some expectations are already set for the Publisher.Publish method")
	}

	mmPublish.mock.funcPublish = f
	return mmPublish.mock
}

// Publish implements cloudmetrics.Publisher
func (mmPublish *PublisherMock) Publish() {
	mm_atomic.AddUint64(&mmPublish.beforePublishCounter, 1)
	defer mm_atomic.AddUint64(&mmPublish.afterPublishCounter, 1)

	if mmPublish.inspectFuncPublish != nil {
		mmPublish.inspectFuncPublish()
	}

	if mmPublish.PublishMock.defaultExpectation != nil {
		mm_atomic.AddUint64(&mmPublish.PublishMock.defaultExpectation.Counter, 1)

		return

	}
	if mmPublish.funcPublish != nil {
		mmPublish.funcPublish()
		return
	}
	mmPublish.t.Fatalf("Unexpected call to PublisherMock.Publish.")

}

// PublishAfterCounter returns a count of finished PublisherMock.Publish invocations
func (mmPublish *PublisherMock) PublishAfterCounter() uint64 {
	return mm_atomic.LoadUint64(&mmPublish.afterPublishCounter)
}

// PublishBeforeCounter returns a count of PublisherMock.Publish invocations
func (mmPublish *PublisherMock) PublishBeforeCounter() uint64 {
	return mm_atomic.LoadUint64(&mmPublish.beforePublishCounter)
}

// MinimockPublishDone returns true if the count of the Publish invocations corresponds
// the number of defined expectations
func (m *PublisherMock) MinimockPublishDone() bool {
	for _, e := range m.PublishMock.expectations {
		if mm_atomic.LoadUint64(&e.Counter) < 1 {
			return false
		}
	}

	// if default expectation was set then invocations count should be greater than zero
	if m.PublishMock.defaultExpectation != nil && mm_atomic.LoadUint64(&m.afterPublishCounter) < 1 {
		return false
	}
	// if func was set then invocations count should be greater than zero
	if m.funcPublish != nil && mm_atomic.LoadUint64(&m.afterPublishCounter) < 1 {
		return false
	}
	return true
}

// MinimockPublishInspect logs each unmet expectation
func (m *PublisherMock) MinimockPublishInspect() {
	for _, e := range m.PublishMock.expectations {
		if mm_atomic.LoadUint64(&e.Counter) < 1 {
			m.t.Error("Expected call to PublisherMock.Publish")
		}
	}

	// if default expectation was set then invocations count should be greater than zero
	if m.PublishMock.defaultExpectation != nil && mm_atomic.LoadUint64(&m.afterPublishCounter) < 1 {
		m.t.Error("Expected call to PublisherMock.Publish")
	}
	// if func was set then invocations count should be greater than zero
	if m.funcPublish != nil && mm_atomic.LoadUint64(&m.afterPublishCounter) < 1 {
		m.t.Error("Expected call to PublisherMock.Publish")
	}
}

// MinimockFinish checks that all mocked methods have been called the expected number of times
func (m *PublisherMock) MinimockFinish() {
	if !m.minimockDone() {
		m.MinimockPublishInspect()
		m.t.FailNow()
	}
}

// MinimockWait waits for all mocked methods to be called the expected number of times
func (m *PublisherMock) MinimockWait(timeout mm_time.Duration) {
	timeoutCh := mm_time.After(timeout)
	for {
		if m.minimockDone() {
			return
		}
		select {
		case <-timeoutCh:
			m.MinimockFinish()
			return
		case <-mm_time.After(10 * mm_time.Millisecond):
		}
	}
}

func (m *PublisherMock) minimockDone() bool {
	done := true
	return done &&
		m.MinimockPublishDone()
}
