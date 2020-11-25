// Code generated by mockery v1.0.0. DO NOT EDIT.

package storage

import (
	context "context"

	storage "github.com/coinbase/rosetta-sdk-go/storage"
	mock "github.com/stretchr/testify/mock"
)

// Database is an autogenerated mock type for the Database type
type Database struct {
	mock.Mock
}

// Close provides a mock function with given fields: _a0
func (_m *Database) Close(_a0 context.Context) error {
	ret := _m.Called(_a0)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context) error); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Encoder provides a mock function with given fields:
func (_m *Database) Encoder() *storage.Encoder {
	ret := _m.Called()

	var r0 *storage.Encoder
	if rf, ok := ret.Get(0).(func() *storage.Encoder); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*storage.Encoder)
		}
	}

	return r0
}

// ReadTransaction provides a mock function with given fields: _a0
func (_m *Database) ReadTransaction(_a0 context.Context) storage.DatabaseTransaction {
	ret := _m.Called(_a0)

	var r0 storage.DatabaseTransaction
	if rf, ok := ret.Get(0).(func(context.Context) storage.DatabaseTransaction); ok {
		r0 = rf(_a0)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(storage.DatabaseTransaction)
		}
	}

	return r0
}

// Transaction provides a mock function with given fields: _a0
func (_m *Database) Transaction(_a0 context.Context) storage.DatabaseTransaction {
	ret := _m.Called(_a0)

	var r0 storage.DatabaseTransaction
	if rf, ok := ret.Get(0).(func(context.Context) storage.DatabaseTransaction); ok {
		r0 = rf(_a0)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(storage.DatabaseTransaction)
		}
	}

	return r0
}

// WriteTransaction provides a mock function with given fields: ctx, identifier, priority
func (_m *Database) WriteTransaction(ctx context.Context, identifier string, priority bool) storage.DatabaseTransaction {
	ret := _m.Called(ctx, identifier, priority)

	var r0 storage.DatabaseTransaction
	if rf, ok := ret.Get(0).(func(context.Context, string, bool) storage.DatabaseTransaction); ok {
		r0 = rf(ctx, identifier, priority)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(storage.DatabaseTransaction)
		}
	}

	return r0
}
