// Code generated by mockery v2.12.2. DO NOT EDIT.

package mocks

import (
	mock "github.com/stretchr/testify/mock"
	model "gitlab.com/project-emco/core/emco-base/src/tac/pkg/model"

	pkgmodule "gitlab.com/project-emco/core/emco-base/src/workflowmgr/pkg/module"

	testing "testing"
)

// WorkflowIntentManager is an autogenerated mock type for the WorkflowIntentManager type
type WorkflowIntentManager struct {
	mock.Mock
}

// CancelWorkflowIntent provides a mock function with given fields: name, project, cApp, cAppVer, dig, req
func (_m *WorkflowIntentManager) CancelWorkflowIntent(name string, project string, cApp string, cAppVer string, dig string, req *model.WfhTemporalCancelRequest) error {
	ret := _m.Called(name, project, cApp, cAppVer, dig, req)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, string, string, string, string, *model.WfhTemporalCancelRequest) error); ok {
		r0 = rf(name, project, cApp, cAppVer, dig, req)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// CreateWorkflowHookIntent provides a mock function with given fields: wfh, project, cApp, cAppVer, dig, exists
func (_m *WorkflowIntentManager) CreateWorkflowHookIntent(wfh model.WorkflowHookIntent, project string, cApp string, cAppVer string, dig string, exists bool) (model.WorkflowHookIntent, error) {
	ret := _m.Called(wfh, project, cApp, cAppVer, dig, exists)

	var r0 model.WorkflowHookIntent
	if rf, ok := ret.Get(0).(func(model.WorkflowHookIntent, string, string, string, string, bool) model.WorkflowHookIntent); ok {
		r0 = rf(wfh, project, cApp, cAppVer, dig, exists)
	} else {
		r0 = ret.Get(0).(model.WorkflowHookIntent)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(model.WorkflowHookIntent, string, string, string, string, bool) error); ok {
		r1 = rf(wfh, project, cApp, cAppVer, dig, exists)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// DeleteWorkflowHookIntent provides a mock function with given fields: name, project, cApp, cAppVer, dig
func (_m *WorkflowIntentManager) DeleteWorkflowHookIntent(name string, project string, cApp string, cAppVer string, dig string) error {
	ret := _m.Called(name, project, cApp, cAppVer, dig)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, string, string, string, string) error); ok {
		r0 = rf(name, project, cApp, cAppVer, dig)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// GetSpecificHooks provides a mock function with given fields: project, cApp, cAppVer, dig, hook
func (_m *WorkflowIntentManager) GetSpecificHooks(project string, cApp string, cAppVer string, dig string, hook string) ([]model.WorkflowHookIntent, error) {
	ret := _m.Called(project, cApp, cAppVer, dig, hook)

	var r0 []model.WorkflowHookIntent
	if rf, ok := ret.Get(0).(func(string, string, string, string, string) []model.WorkflowHookIntent); ok {
		r0 = rf(project, cApp, cAppVer, dig, hook)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]model.WorkflowHookIntent)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, string, string, string, string) error); ok {
		r1 = rf(project, cApp, cAppVer, dig, hook)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetStatusWorkflowIntent provides a mock function with given fields: name, project, cApp, cAppVer, dig, query
func (_m *WorkflowIntentManager) GetStatusWorkflowIntent(name string, project string, cApp string, cAppVer string, dig string, query *pkgmodule.WfTemporalStatusQuery) (*pkgmodule.WfTemporalStatusResponse, error) {
	ret := _m.Called(name, project, cApp, cAppVer, dig, query)

	var r0 *pkgmodule.WfTemporalStatusResponse
	if rf, ok := ret.Get(0).(func(string, string, string, string, string, *pkgmodule.WfTemporalStatusQuery) *pkgmodule.WfTemporalStatusResponse); ok {
		r0 = rf(name, project, cApp, cAppVer, dig, query)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*pkgmodule.WfTemporalStatusResponse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, string, string, string, string, *pkgmodule.WfTemporalStatusQuery) error); ok {
		r1 = rf(name, project, cApp, cAppVer, dig, query)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetWorkflowHookIntent provides a mock function with given fields: name, project, cApp, cAppVer, dig
func (_m *WorkflowIntentManager) GetWorkflowHookIntent(name string, project string, cApp string, cAppVer string, dig string) (model.WorkflowHookIntent, error) {
	ret := _m.Called(name, project, cApp, cAppVer, dig)

	var r0 model.WorkflowHookIntent
	if rf, ok := ret.Get(0).(func(string, string, string, string, string) model.WorkflowHookIntent); ok {
		r0 = rf(name, project, cApp, cAppVer, dig)
	} else {
		r0 = ret.Get(0).(model.WorkflowHookIntent)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, string, string, string, string) error); ok {
		r1 = rf(name, project, cApp, cAppVer, dig)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetWorkflowHookIntents provides a mock function with given fields: project, cApp, cAppVer, dig
func (_m *WorkflowIntentManager) GetWorkflowHookIntents(project string, cApp string, cAppVer string, dig string) ([]model.WorkflowHookIntent, error) {
	ret := _m.Called(project, cApp, cAppVer, dig)

	var r0 []model.WorkflowHookIntent
	if rf, ok := ret.Get(0).(func(string, string, string, string) []model.WorkflowHookIntent); ok {
		r0 = rf(project, cApp, cAppVer, dig)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]model.WorkflowHookIntent)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, string, string, string) error); ok {
		r1 = rf(project, cApp, cAppVer, dig)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// NewWorkflowIntentManager creates a new instance of WorkflowIntentManager. It also registers the testing.TB interface on the mock and a cleanup function to assert the mocks expectations.
func NewWorkflowIntentManager(t testing.TB) *WorkflowIntentManager {
	mock := &WorkflowIntentManager{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}