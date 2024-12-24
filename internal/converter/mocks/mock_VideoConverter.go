// Code generated by mockery v2.50.0. DO NOT EDIT.

package mocks

import (
	converter "github.com/chistyakoviv/converter/internal/converter"
	mock "github.com/stretchr/testify/mock"
)

// MockVideoConverter is an autogenerated mock type for the VideoConverter type
type MockVideoConverter struct {
	mock.Mock
}

type MockVideoConverter_Expecter struct {
	mock *mock.Mock
}

func (_m *MockVideoConverter) EXPECT() *MockVideoConverter_Expecter {
	return &MockVideoConverter_Expecter{mock: &_m.Mock}
}

// Convert provides a mock function with given fields: from, to, conf
func (_m *MockVideoConverter) Convert(from string, to string, conf converter.ConversionConfig) error {
	ret := _m.Called(from, to, conf)

	if len(ret) == 0 {
		panic("no return value specified for Convert")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(string, string, converter.ConversionConfig) error); ok {
		r0 = rf(from, to, conf)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// MockVideoConverter_Convert_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Convert'
type MockVideoConverter_Convert_Call struct {
	*mock.Call
}

// Convert is a helper method to define mock.On call
//   - from string
//   - to string
//   - conf converter.ConversionConfig
func (_e *MockVideoConverter_Expecter) Convert(from interface{}, to interface{}, conf interface{}) *MockVideoConverter_Convert_Call {
	return &MockVideoConverter_Convert_Call{Call: _e.mock.On("Convert", from, to, conf)}
}

func (_c *MockVideoConverter_Convert_Call) Run(run func(from string, to string, conf converter.ConversionConfig)) *MockVideoConverter_Convert_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string), args[1].(string), args[2].(converter.ConversionConfig))
	})
	return _c
}

func (_c *MockVideoConverter_Convert_Call) Return(_a0 error) *MockVideoConverter_Convert_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockVideoConverter_Convert_Call) RunAndReturn(run func(string, string, converter.ConversionConfig) error) *MockVideoConverter_Convert_Call {
	_c.Call.Return(run)
	return _c
}

// Shutdown provides a mock function with no fields
func (_m *MockVideoConverter) Shutdown() {
	_m.Called()
}

// MockVideoConverter_Shutdown_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Shutdown'
type MockVideoConverter_Shutdown_Call struct {
	*mock.Call
}

// Shutdown is a helper method to define mock.On call
func (_e *MockVideoConverter_Expecter) Shutdown() *MockVideoConverter_Shutdown_Call {
	return &MockVideoConverter_Shutdown_Call{Call: _e.mock.On("Shutdown")}
}

func (_c *MockVideoConverter_Shutdown_Call) Run(run func()) *MockVideoConverter_Shutdown_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockVideoConverter_Shutdown_Call) Return() *MockVideoConverter_Shutdown_Call {
	_c.Call.Return()
	return _c
}

func (_c *MockVideoConverter_Shutdown_Call) RunAndReturn(run func()) *MockVideoConverter_Shutdown_Call {
	_c.Run(run)
	return _c
}

// NewMockVideoConverter creates a new instance of MockVideoConverter. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockVideoConverter(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockVideoConverter {
	mock := &MockVideoConverter{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}