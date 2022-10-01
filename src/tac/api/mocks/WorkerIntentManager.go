// Code generated by mockery v2.12.2. DO NOT EDIT.

package mocks

import (
	mock "github.com/stretchr/testify/mock"
	model "gitlab.com/project-emco/core/emco-base/src/tac/pkg/model"

	testing "testing"
)

// WorkerIntentManager is an autogenerated mock type for the WorkerIntentManager type
type WorkerIntentManager struct {
	mock.Mock
}

// CreateOrUpdateWorkerIntent provides a mock function with given fields: wi, tac, project, cApp, cAppVer, dig, exists
func (_m *WorkerIntentManager) CreateOrUpdateWorkerIntent(wi model.WorkerIntent, tac string, project string, cApp string, cAppVer string, dig string, exists bool) (model.WorkerIntent, error) {
	ret := _m.Called(wi, tac, project, cApp, cAppVer, dig, exists)

	var r0 model.WorkerIntent
	if rf, ok := ret.Get(0).(func(model.WorkerIntent, string, string, string, string, string, bool) model.WorkerIntent); ok {
		r0 = rf(wi, tac, project, cApp, cAppVer, dig, exists)
	} else {
		r0 = ret.Get(0).(model.WorkerIntent)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(model.WorkerIntent, string, string, string, string, string, bool) error); ok {
		r1 = rf(wi, tac, project, cApp, cAppVer, dig, exists)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// DeleteWorkerIntents provides a mock function with given fields: project, cApp, cAppVer, dig, tac, workerName
func (_m *WorkerIntentManager) DeleteWorkerIntents(project string, cApp string, cAppVer string, dig string, tac string, workerName string) error {
	ret := _m.Called(project, cApp, cAppVer, dig, tac, workerName)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, string, string, string, string, string) error); ok {
		r0 = rf(project, cApp, cAppVer, dig, tac, workerName)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// GetWorkerIntent provides a mock function with given fields: workerName, project, cApp, cAppVer, dig, tac
func (_m *WorkerIntentManager) GetWorkerIntent(workerName string, project string, cApp string, cAppVer string, dig string, tac string) (model.WorkerIntent, error) {
	ret := _m.Called(workerName, project, cApp, cAppVer, dig, tac)

	var r0 model.WorkerIntent
	if rf, ok := ret.Get(0).(func(string, string, string, string, string, string) model.WorkerIntent); ok {
		r0 = rf(workerName, project, cApp, cAppVer, dig, tac)
	} else {
		r0 = ret.Get(0).(model.WorkerIntent)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, string, string, string, string, string) error); ok {
		r1 = rf(workerName, project, cApp, cAppVer, dig, tac)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetWorkerIntents provides a mock function with given fields: project, cApp, cAppVer, dig, tac
func (_m *WorkerIntentManager) GetWorkerIntents(project string, cApp string, cAppVer string, dig string, tac string) ([]model.WorkerIntent, error) {
	ret := _m.Called(project, cApp, cAppVer, dig, tac)

	var r0 []model.WorkerIntent
	if rf, ok := ret.Get(0).(func(string, string, string, string, string) []model.WorkerIntent); ok {
		r0 = rf(project, cApp, cAppVer, dig, tac)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]model.WorkerIntent)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, string, string, string, string) error); ok {
		r1 = rf(project, cApp, cAppVer, dig, tac)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// NewWorkerIntentManager creates a new instance of WorkerIntentManager. It also registers the testing.TB interface on the mock and a cleanup function to assert the mocks expectations.
func NewWorkerIntentManager(t testing.TB) *WorkerIntentManager {
	mock := &WorkerIntentManager{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}