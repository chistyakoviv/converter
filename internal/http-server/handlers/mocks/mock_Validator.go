// Code generated by mockery v2.50.0. DO NOT EDIT.

package mocks

import mock "github.com/stretchr/testify/mock"

// MockValidator is an autogenerated mock type for the Validator type
type MockValidator struct {
	mock.Mock
}

type MockValidator_Expecter struct {
	mock *mock.Mock
}

func (_m *MockValidator) EXPECT() *MockValidator_Expecter {
	return &MockValidator_Expecter{mock: &_m.Mock}
}

// Struct provides a mock function with given fields: s
func (_m *MockValidator) Struct(s interface{}) error {
	ret := _m.Called(s)

	if len(ret) == 0 {
		panic("no return value specified for Struct")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(interface{}) error); ok {
		r0 = rf(s)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// MockValidator_Struct_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Struct'
type MockValidator_Struct_Call struct {
	*mock.Call
}

// Struct is a helper method to define mock.On call
//   - s interface{}
func (_e *MockValidator_Expecter) Struct(s interface{}) *MockValidator_Struct_Call {
	return &MockValidator_Struct_Call{Call: _e.mock.On("Struct", s)}
}

func (_c *MockValidator_Struct_Call) Run(run func(s interface{})) *MockValidator_Struct_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(interface{}))
	})
	return _c
}

func (_c *MockValidator_Struct_Call) Return(_a0 error) *MockValidator_Struct_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockValidator_Struct_Call) RunAndReturn(run func(interface{}) error) *MockValidator_Struct_Call {
	_c.Call.Return(run)
	return _c
}

// NewMockValidator creates a new instance of MockValidator. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockValidator(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockValidator {
	mock := &MockValidator{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}