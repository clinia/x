// Code generated by mockery v2.44.1. DO NOT EDIT.

package mocks

import (
	context "context"

	messagex "github.com/clinia/x/pubsubx/messagex"
	mock "github.com/stretchr/testify/mock"

	pubsubx "github.com/clinia/x/pubsubx"
)

// Publisher is an autogenerated mock type for the Publisher type
type Publisher struct {
	mock.Mock
}

// Close provides a mock function with given fields:
func (_m *Publisher) Close() error {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Close")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// PublishAsync provides a mock function with given fields: ctx, topic, messages
func (_m *Publisher) PublishAsync(ctx context.Context, topic messagex.Topic, messages ...*messagex.Message) error {
	_va := make([]interface{}, len(messages))
	for _i := range messages {
		_va[_i] = messages[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, topic)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for PublishAsync")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, messagex.Topic, ...*messagex.Message) error); ok {
		r0 = rf(ctx, topic, messages...)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// PublishSync provides a mock function with given fields: ctx, topic, messages
func (_m *Publisher) PublishSync(ctx context.Context, topic messagex.Topic, messages ...*messagex.Message) (pubsubx.Errors, error) {
	_va := make([]interface{}, len(messages))
	for _i := range messages {
		_va[_i] = messages[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, topic)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for PublishSync")
	}

	var r0 pubsubx.Errors
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, messagex.Topic, ...*messagex.Message) (pubsubx.Errors, error)); ok {
		return rf(ctx, topic, messages...)
	}
	if rf, ok := ret.Get(0).(func(context.Context, messagex.Topic, ...*messagex.Message) pubsubx.Errors); ok {
		r0 = rf(ctx, topic, messages...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(pubsubx.Errors)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, messagex.Topic, ...*messagex.Message) error); ok {
		r1 = rf(ctx, topic, messages...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// NewPublisher creates a new instance of Publisher. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewPublisher(t interface {
	mock.TestingT
	Cleanup(func())
}) *Publisher {
	mock := &Publisher{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
