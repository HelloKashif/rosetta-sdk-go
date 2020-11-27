// Code generated by mockery v1.0.0. DO NOT EDIT.

package syncer

import (
	context "context"

	mock "github.com/stretchr/testify/mock"

	types "github.com/HelloKashif/rosetta-sdk-go/types"
)

// Handler is an autogenerated mock type for the Handler type
type Handler struct {
	mock.Mock
}

// BlockAdded provides a mock function with given fields: ctx, block
func (_m *Handler) BlockAdded(ctx context.Context, block *types.Block) error {
	ret := _m.Called(ctx, block)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *types.Block) error); ok {
		r0 = rf(ctx, block)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// BlockRemoved provides a mock function with given fields: ctx, block
func (_m *Handler) BlockRemoved(ctx context.Context, block *types.BlockIdentifier) error {
	ret := _m.Called(ctx, block)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *types.BlockIdentifier) error); ok {
		r0 = rf(ctx, block)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
