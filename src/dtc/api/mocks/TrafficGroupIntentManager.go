// Code generated by mockery v2.5.1. DO NOT EDIT.

package mocks

import (
	"context"
	module "gitlab.com/project-emco/core/emco-base/src/dtc/pkg/module"
	mock "github.com/stretchr/testify/mock"
)

// TrafficGroupIntentManager is an autogenerated mock type for the TrafficGroupIntentManager type
type TrafficGroupIntentManager struct {
	mock.Mock
}

// CreateTrafficGroupIntent provides a mock function with given fields: tci, project, compositeapp, compositeappversion, deploymentIntentGroupName, exists
func (_m *TrafficGroupIntentManager) CreateTrafficGroupIntent(ctx context.Context, tci module.TrafficGroupIntent, project string, compositeapp string, compositeappversion string, deploymentIntentGroupName string, exists bool) (module.TrafficGroupIntent, error) {
	ret := _m.Called(ctx, tci, project, compositeapp, compositeappversion, deploymentIntentGroupName, exists)

	var r0 module.TrafficGroupIntent
	if rf, ok := ret.Get(0).(func(module.TrafficGroupIntent, string, string, string, string, bool) module.TrafficGroupIntent); ok {
		r0 = rf(tci, project, compositeapp, compositeappversion, deploymentIntentGroupName, exists)
	} else {
		r0 = ret.Get(0).(module.TrafficGroupIntent)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(module.TrafficGroupIntent, string, string, string, string, bool) error); ok {
		r1 = rf(tci, project, compositeapp, compositeappversion, deploymentIntentGroupName, exists)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// DeleteTrafficGroupIntent provides a mock function with given fields: name, project, compositeapp, compositeappversion, dig
func (_m *TrafficGroupIntentManager) DeleteTrafficGroupIntent(ctx context.Context, name string, project string, compositeapp string, compositeappversion string, dig string) error {
	ret := _m.Called(ctx, name, project, compositeapp, compositeappversion, dig)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, string, string, string, string) error); ok {
		r0 = rf(name, project, compositeapp, compositeappversion, dig)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// GetTrafficGroupIntent provides a mock function with given fields: name, project, compositeapp, compositeappversion, dig
func (_m *TrafficGroupIntentManager) GetTrafficGroupIntent(ctx context.Context, name string, project string, compositeapp string, compositeappversion string, dig string) (module.TrafficGroupIntent, error) {
	ret := _m.Called(ctx, name, project, compositeapp, compositeappversion, dig)

	var r0 module.TrafficGroupIntent
	if rf, ok := ret.Get(0).(func(string, string, string, string, string) module.TrafficGroupIntent); ok {
		r0 = rf(name, project, compositeapp, compositeappversion, dig)
	} else {
		r0 = ret.Get(0).(module.TrafficGroupIntent)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, string, string, string, string) error); ok {
		r1 = rf(name, project, compositeapp, compositeappversion, dig)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetTrafficGroupIntents provides a mock function with given fields: project, compositeapp, compositeappversion, dig
func (_m *TrafficGroupIntentManager) GetTrafficGroupIntents(ctx context.Context, project string, compositeapp string, compositeappversion string, dig string) ([]module.TrafficGroupIntent, error) {
	ret := _m.Called(ctx, project, compositeapp, compositeappversion, dig)

	var r0 []module.TrafficGroupIntent
	if rf, ok := ret.Get(0).(func(string, string, string, string) []module.TrafficGroupIntent); ok {
		r0 = rf(project, compositeapp, compositeappversion, dig)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]module.TrafficGroupIntent)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, string, string, string) error); ok {
		r1 = rf(project, compositeapp, compositeappversion, dig)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
