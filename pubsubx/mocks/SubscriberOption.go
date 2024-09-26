// Code generated by mockery v2.44.1. DO NOT EDIT.

package mocks

import (
	pubsubx "github.com/clinia/x/pubsubx"
	mock "github.com/stretchr/testify/mock"
)

// SubscriberOption is an autogenerated mock type for the SubscriberOption type
type SubscriberOption struct {
	mock.Mock
}

// Execute provides a mock function with given fields: _a0
func (_m *SubscriberOption) Execute(_a0 *pubsubx.SubscriberOptions) {
	_m.Called(_a0)
}

// NewSubscriberOption creates a new instance of SubscriberOption. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewSubscriberOption(t interface {
	mock.TestingT
	Cleanup(func())
}) *SubscriberOption {
	mock := &SubscriberOption{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}