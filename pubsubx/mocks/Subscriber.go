// Code generated by mockery v2.44.1. DO NOT EDIT.

package mocks

import (
	context "context"

	pubsubx "github.com/clinia/x/pubsubx"
	mock "github.com/stretchr/testify/mock"
)

// Subscriber is an autogenerated mock type for the Subscriber type
type Subscriber struct {
	mock.Mock
}

// Close provides a mock function with given fields:
func (_m *Subscriber) Close() error {
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

// Health provides a mock function with given fields:
func (_m *Subscriber) Health() error {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Health")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Subscribe provides a mock function with given fields: ctx, topicHandlers
func (_m *Subscriber) Subscribe(ctx context.Context, topicHandlers pubsubx.Handlers) error {
	ret := _m.Called(ctx, topicHandlers)

	if len(ret) == 0 {
		panic("no return value specified for Subscribe")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, pubsubx.Handlers) error); ok {
		r0 = rf(ctx, topicHandlers)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// NewSubscriber creates a new instance of Subscriber. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewSubscriber(t interface {
	mock.TestingT
	Cleanup(func())
}) *Subscriber {
	mock := &Subscriber{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
