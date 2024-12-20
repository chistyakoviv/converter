// Code generated by mockery v2.50.0. DO NOT EDIT.

package mocks

import (
	context "context"

	model "github.com/chistyakoviv/converter/internal/model"
	mock "github.com/stretchr/testify/mock"
)

// MockDeletionQueueRepository is an autogenerated mock type for the DeletionQueueRepository type
type MockDeletionQueueRepository struct {
	mock.Mock
}

type MockDeletionQueueRepository_Expecter struct {
	mock *mock.Mock
}

func (_m *MockDeletionQueueRepository) EXPECT() *MockDeletionQueueRepository_Expecter {
	return &MockDeletionQueueRepository_Expecter{mock: &_m.Mock}
}

// Create provides a mock function with given fields: ctx, file
func (_m *MockDeletionQueueRepository) Create(ctx context.Context, file *model.DeletionInfo) (int64, error) {
	ret := _m.Called(ctx, file)

	if len(ret) == 0 {
		panic("no return value specified for Create")
	}

	var r0 int64
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *model.DeletionInfo) (int64, error)); ok {
		return rf(ctx, file)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *model.DeletionInfo) int64); ok {
		r0 = rf(ctx, file)
	} else {
		r0 = ret.Get(0).(int64)
	}

	if rf, ok := ret.Get(1).(func(context.Context, *model.DeletionInfo) error); ok {
		r1 = rf(ctx, file)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockDeletionQueueRepository_Create_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Create'
type MockDeletionQueueRepository_Create_Call struct {
	*mock.Call
}

// Create is a helper method to define mock.On call
//   - ctx context.Context
//   - file *model.DeletionInfo
func (_e *MockDeletionQueueRepository_Expecter) Create(ctx interface{}, file interface{}) *MockDeletionQueueRepository_Create_Call {
	return &MockDeletionQueueRepository_Create_Call{Call: _e.mock.On("Create", ctx, file)}
}

func (_c *MockDeletionQueueRepository_Create_Call) Run(run func(ctx context.Context, file *model.DeletionInfo)) *MockDeletionQueueRepository_Create_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*model.DeletionInfo))
	})
	return _c
}

func (_c *MockDeletionQueueRepository_Create_Call) Return(_a0 int64, _a1 error) *MockDeletionQueueRepository_Create_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockDeletionQueueRepository_Create_Call) RunAndReturn(run func(context.Context, *model.DeletionInfo) (int64, error)) *MockDeletionQueueRepository_Create_Call {
	_c.Call.Return(run)
	return _c
}

// FindByFullpath provides a mock function with given fields: ctx, fullpath
func (_m *MockDeletionQueueRepository) FindByFullpath(ctx context.Context, fullpath string) (*model.Deletion, error) {
	ret := _m.Called(ctx, fullpath)

	if len(ret) == 0 {
		panic("no return value specified for FindByFullpath")
	}

	var r0 *model.Deletion
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) (*model.Deletion, error)); ok {
		return rf(ctx, fullpath)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) *model.Deletion); ok {
		r0 = rf(ctx, fullpath)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.Deletion)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, fullpath)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockDeletionQueueRepository_FindByFullpath_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'FindByFullpath'
type MockDeletionQueueRepository_FindByFullpath_Call struct {
	*mock.Call
}

// FindByFullpath is a helper method to define mock.On call
//   - ctx context.Context
//   - fullpath string
func (_e *MockDeletionQueueRepository_Expecter) FindByFullpath(ctx interface{}, fullpath interface{}) *MockDeletionQueueRepository_FindByFullpath_Call {
	return &MockDeletionQueueRepository_FindByFullpath_Call{Call: _e.mock.On("FindByFullpath", ctx, fullpath)}
}

func (_c *MockDeletionQueueRepository_FindByFullpath_Call) Run(run func(ctx context.Context, fullpath string)) *MockDeletionQueueRepository_FindByFullpath_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string))
	})
	return _c
}

func (_c *MockDeletionQueueRepository_FindByFullpath_Call) Return(_a0 *model.Deletion, _a1 error) *MockDeletionQueueRepository_FindByFullpath_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockDeletionQueueRepository_FindByFullpath_Call) RunAndReturn(run func(context.Context, string) (*model.Deletion, error)) *MockDeletionQueueRepository_FindByFullpath_Call {
	_c.Call.Return(run)
	return _c
}

// FindOldestQueued provides a mock function with given fields: ctx
func (_m *MockDeletionQueueRepository) FindOldestQueued(ctx context.Context) (*model.Deletion, error) {
	ret := _m.Called(ctx)

	if len(ret) == 0 {
		panic("no return value specified for FindOldestQueued")
	}

	var r0 *model.Deletion
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context) (*model.Deletion, error)); ok {
		return rf(ctx)
	}
	if rf, ok := ret.Get(0).(func(context.Context) *model.Deletion); ok {
		r0 = rf(ctx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.Deletion)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = rf(ctx)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockDeletionQueueRepository_FindOldestQueued_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'FindOldestQueued'
type MockDeletionQueueRepository_FindOldestQueued_Call struct {
	*mock.Call
}

// FindOldestQueued is a helper method to define mock.On call
//   - ctx context.Context
func (_e *MockDeletionQueueRepository_Expecter) FindOldestQueued(ctx interface{}) *MockDeletionQueueRepository_FindOldestQueued_Call {
	return &MockDeletionQueueRepository_FindOldestQueued_Call{Call: _e.mock.On("FindOldestQueued", ctx)}
}

func (_c *MockDeletionQueueRepository_FindOldestQueued_Call) Run(run func(ctx context.Context)) *MockDeletionQueueRepository_FindOldestQueued_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context))
	})
	return _c
}

func (_c *MockDeletionQueueRepository_FindOldestQueued_Call) Return(_a0 *model.Deletion, _a1 error) *MockDeletionQueueRepository_FindOldestQueued_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockDeletionQueueRepository_FindOldestQueued_Call) RunAndReturn(run func(context.Context) (*model.Deletion, error)) *MockDeletionQueueRepository_FindOldestQueued_Call {
	_c.Call.Return(run)
	return _c
}

// MarkAsCanceled provides a mock function with given fields: ctx, fullpath, code
func (_m *MockDeletionQueueRepository) MarkAsCanceled(ctx context.Context, fullpath string, code uint32) error {
	ret := _m.Called(ctx, fullpath, code)

	if len(ret) == 0 {
		panic("no return value specified for MarkAsCanceled")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, uint32) error); ok {
		r0 = rf(ctx, fullpath, code)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// MockDeletionQueueRepository_MarkAsCanceled_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'MarkAsCanceled'
type MockDeletionQueueRepository_MarkAsCanceled_Call struct {
	*mock.Call
}

// MarkAsCanceled is a helper method to define mock.On call
//   - ctx context.Context
//   - fullpath string
//   - code uint32
func (_e *MockDeletionQueueRepository_Expecter) MarkAsCanceled(ctx interface{}, fullpath interface{}, code interface{}) *MockDeletionQueueRepository_MarkAsCanceled_Call {
	return &MockDeletionQueueRepository_MarkAsCanceled_Call{Call: _e.mock.On("MarkAsCanceled", ctx, fullpath, code)}
}

func (_c *MockDeletionQueueRepository_MarkAsCanceled_Call) Run(run func(ctx context.Context, fullpath string, code uint32)) *MockDeletionQueueRepository_MarkAsCanceled_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string), args[2].(uint32))
	})
	return _c
}

func (_c *MockDeletionQueueRepository_MarkAsCanceled_Call) Return(_a0 error) *MockDeletionQueueRepository_MarkAsCanceled_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockDeletionQueueRepository_MarkAsCanceled_Call) RunAndReturn(run func(context.Context, string, uint32) error) *MockDeletionQueueRepository_MarkAsCanceled_Call {
	_c.Call.Return(run)
	return _c
}

// MarkAsDone provides a mock function with given fields: ctx, fullpath
func (_m *MockDeletionQueueRepository) MarkAsDone(ctx context.Context, fullpath string) error {
	ret := _m.Called(ctx, fullpath)

	if len(ret) == 0 {
		panic("no return value specified for MarkAsDone")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string) error); ok {
		r0 = rf(ctx, fullpath)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// MockDeletionQueueRepository_MarkAsDone_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'MarkAsDone'
type MockDeletionQueueRepository_MarkAsDone_Call struct {
	*mock.Call
}

// MarkAsDone is a helper method to define mock.On call
//   - ctx context.Context
//   - fullpath string
func (_e *MockDeletionQueueRepository_Expecter) MarkAsDone(ctx interface{}, fullpath interface{}) *MockDeletionQueueRepository_MarkAsDone_Call {
	return &MockDeletionQueueRepository_MarkAsDone_Call{Call: _e.mock.On("MarkAsDone", ctx, fullpath)}
}

func (_c *MockDeletionQueueRepository_MarkAsDone_Call) Run(run func(ctx context.Context, fullpath string)) *MockDeletionQueueRepository_MarkAsDone_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string))
	})
	return _c
}

func (_c *MockDeletionQueueRepository_MarkAsDone_Call) Return(_a0 error) *MockDeletionQueueRepository_MarkAsDone_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockDeletionQueueRepository_MarkAsDone_Call) RunAndReturn(run func(context.Context, string) error) *MockDeletionQueueRepository_MarkAsDone_Call {
	_c.Call.Return(run)
	return _c
}

// NewMockDeletionQueueRepository creates a new instance of MockDeletionQueueRepository. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockDeletionQueueRepository(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockDeletionQueueRepository {
	mock := &MockDeletionQueueRepository{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}