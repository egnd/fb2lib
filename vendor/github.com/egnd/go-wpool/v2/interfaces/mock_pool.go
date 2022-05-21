// Code generated by mockery v2.12.0. DO NOT EDIT.

package interfaces

import (
	testing "testing"

	mock "github.com/stretchr/testify/mock"
)

// MockPool is an autogenerated mock type for the Pool type
type MockPool struct {
	mock.Mock
}

// AddTask provides a mock function with given fields: _a0
func (_m *MockPool) AddTask(_a0 Task) error {
	ret := _m.Called(_a0)

	var r0 error
	if rf, ok := ret.Get(0).(func(Task) error); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// AddWorker provides a mock function with given fields: _a0
func (_m *MockPool) AddWorker(_a0 Worker) error {
	ret := _m.Called(_a0)

	var r0 error
	if rf, ok := ret.Get(0).(func(Worker) error); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Close provides a mock function with given fields:
func (_m *MockPool) Close() error {
	ret := _m.Called()

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// NewMockPool creates a new instance of MockPool. It also registers the testing.TB interface on the mock and a cleanup function to assert the mocks expectations.
func NewMockPool(t testing.TB) *MockPool {
	mock := &MockPool{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
