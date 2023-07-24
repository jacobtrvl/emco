// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package module

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	pkgerrors "github.com/pkg/errors"

	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/appcontext"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/db"
	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/state"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/utils"
)

type serviceActionValue string

const (
	StatusCreated      = "Created"
	StatusProcessing   = "Processing"
	StatusInstantiated = "Instantiated"
	StatusTerminated   = "Terminated"
	StatusFailed       = "Failed"

	StatusReady    = "Ready"
	StatusNotReady = "NotReady"

	ActionCreated       serviceActionValue = "Created"
	ActionInstantiating serviceActionValue = "Instantiating"
	ActionTerminating   serviceActionValue = "Terminating"
)

// ServiceRequest Rest request struct
type ServiceRequest struct {
	MetaData ServiceRequestMetaData `json:"metadata"`
	Spec     ServiceSpec            `json:"spec"`
}

// ServiceRequestMetaData metadata of the request Service
type ServiceRequestMetaData struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// Service represents a logical group of DIGs
type Service struct {
	MetaData ServiceMetaData `json:"metadata"`
	Spec     ServiceSpec     `json:"spec"`
}

// ServiceState represents the service action records
type ServiceState struct {
	ServiceActions     []ServiceAction `json:"serviceActions"`
	ToInstantiatedDIGs map[string]bool `json:"toInstantiatedDIGs"`
}

// ServiceMetaData metadata of the Service
type ServiceMetaData struct {
	Name        string `json:"name"`
	Id          string `json:"id"`
	Project     string `json:"project"`
	Description string `json:"description"`
}

// ServiceSpec Spec info of the service
type ServiceSpec struct {
	Digs []string `json:"digs"`
}

type ServiceAction struct {
	State     serviceActionValue `json:"state"`
	TimeStamp time.Time          `json:"time"`
}

func (r *ServiceRequest) ToService() *Service {
	return &Service{
		MetaData: ServiceMetaData{
			Name:        r.MetaData.Name,
			Description: r.MetaData.Description,
		},
		Spec: r.Spec,
	}
}

func (s *Service) verifyContainDigs(digs []string) error {
	if digs == nil {
		return nil
	}

	serviceDigs := s.Spec.Digs
	if serviceDigs == nil {
		serviceDigs = []string{}
	}

	digsDiff := utils.ListDifference(digs, serviceDigs)
	if len(digsDiff) > 0 {
		return fmt.Errorf("DIGs \"%s\" not found in service \"%s\"", digsDiff, s.MetaData.Name)
	}

	return nil
}

func (s *ServiceState) AddState(action serviceActionValue) {
	s.ServiceActions = append(s.ServiceActions, ServiceAction{State: action, TimeStamp: time.Now()})
}

func (s *ServiceState) LastState() *ServiceAction {
	if s.ServiceActions == nil || len(s.ServiceActions) == 0 {
		return nil
	}

	return &s.ServiceActions[len(s.ServiceActions)-1]
}

func (s *ServiceState) IsToInstantiateDIG(digId string) bool {
	_, ok := s.ToInstantiatedDIGs[digId]
	return ok
}

func (s *ServiceState) AddToInstantiateDIGs(digIds []string) {
	for _, digId := range digIds {
		s.ToInstantiatedDIGs[digId] = true
	}
}

func (s *ServiceState) RemoveToInstantiateDIGs(digIds []string) {
	for _, dig := range digIds {
		delete(s.ToInstantiatedDIGs, dig)
	}
}

func (s *ServiceState) ClearToInstantiateDIG() {
	s.ToInstantiatedDIGs = map[string]bool{}
	s.AddState(ActionTerminating)
}

// ServiceDigsUpdate contains update info for adding/removing digs to/from Service
type ServiceDigsUpdate struct {
	Add    []string `json:"add"`
	Remove []string `json:"remove"`
}

// Validate verify that same dig not appearing in "add" and "remove"
// and DIGs Ids are valid
func (sdu *ServiceDigsUpdate) Validate() error {
	if sdu.Add != nil {
		if utils.HasDuplication(sdu.Add) {
			return fmt.Errorf("invalid ServiceDigUpdate, \"add\" has duplication")
		}

		if err := ValidateDigIds(sdu.Add); err != nil {
			return err
		}
	}

	if sdu.Remove != nil {
		if utils.HasDuplication(sdu.Remove) {
			return fmt.Errorf("invalid ServiceDigUpdate, \"remove\" has duplication")
		}

		if err := ValidateDigIds(sdu.Remove); err != nil {
			return err
		}
	}

	if sdu.Add == nil || sdu.Remove == nil {
		return nil
	}

	if utils.HasIntersection(sdu.Add, sdu.Remove) {
		return fmt.Errorf("invalid ServiceDigUpdate, \"add\" and \"remove\" shouldn't have intersection")
	}

	return nil
}

type ServiceDigsAction struct {
	Digs  []string `json:"digs"`
	Force bool     `json:"force"`
}

// ServiceStatusInfo status of the service and DIGs
type ServiceStatusInfo struct {
	DeployedStatus string                      `json:"deployedStatus"`
	ReadyStatus    string                      `json:"readyStatus"`
	DigsStatus     map[string]DeploymentStatus `json:"digsStatus"`
}

// ServiceKey is the key structure that is used in the database
type ServiceKey struct {
	Name    string `json:"service"`
	Project string `json:"project"`
}

// Json marshalling to convert to string to
// preserve the underlying structure.
func (sk ServiceKey) String() string {
	out, err := json.Marshal(sk)
	if err != nil {
		return ""
	}

	return string(out)
}

// ServiceManager is an interface which exposes the ServiceManager functionality
type ServiceManager interface {
	CreateService(ctx context.Context, service *Service) error
	UpdateService(ctx context.Context, service *Service) error
	GetAllServices(ctx context.Context, project string) ([]*Service, error)
	GetService(ctx context.Context, key *ServiceKey) (*Service, error)
	IsServiceExists(ctx context.Context, key *ServiceKey) (bool, error)
	UpdateServiceDigs(ctx context.Context, key *ServiceKey, sdu *ServiceDigsUpdate) (*Service, error)
	DeleteService(ctx context.Context, key *ServiceKey) error
	InstantiateService(ctx context.Context, key *ServiceKey, sda *ServiceDigsAction) error
	TerminateService(ctx context.Context, key *ServiceKey, sda *ServiceDigsAction) error
	InstantiateServiceDIGs(ctx context.Context, key *ServiceKey, sda *ServiceDigsAction) error
	TerminateServiceDIGs(ctx context.Context, key *ServiceKey, sda *ServiceDigsAction) error
	ServiceStatus(ctx context.Context, key *ServiceKey) (*ServiceStatusInfo, error)
}

// ServiceClient implements the ServiceManager interface
type ServiceClient struct {
	storeName           string
	tag                 string
	stateTag            string
	digClient           *DeploymentIntentGroupClient
	InstantiationClient *InstantiationClient
}

// NewServiceClient return an instance of ServiceClient which implements ServiceManager
func NewServiceClient() ServiceManager {
	return &ServiceClient{
		storeName:           "resources",
		tag:                 "data",
		stateTag:            "stateInfo",
		digClient:           NewDeploymentIntentGroupClient(),
		InstantiationClient: NewInstantiationClient(),
	}
}

func (sc *ServiceClient) getDigById(ctx context.Context, digId string) (*DeploymentIntentGroup, error) {
	digKey, err := DeploymentIntentGroupKeyFromDigId(digId)
	if err != nil {
		return nil, err
	}

	dig, err := sc.digClient.GetDeploymentIntentGroup(ctx, digKey.Name, digKey.Project, digKey.CompositeApp, digKey.Version)
	if err != nil {
		return nil, err
	}

	return &dig, nil
}

func (sc *ServiceClient) storeDig(ctx context.Context, digKey *DeploymentIntentGroupKey, dig *DeploymentIntentGroup) error {
	return db.DBconn.Insert(ctx, sc.digClient.storeName, *digKey, nil, sc.tag, dig)
}

func (sc *ServiceClient) storeService(ctx context.Context, service *Service) error {
	key := ServiceKey{
		Name:    service.MetaData.Name,
		Project: service.MetaData.Project,
	}

	return db.DBconn.Insert(ctx, sc.storeName, key, nil, sc.tag, service)
}

// GetServiceState returns the ServiceState records
func (sc *ServiceClient) GetServiceState(ctx context.Context, key *ServiceKey) (*ServiceState, error) {
	result, err := db.DBconn.Find(ctx, sc.storeName, key, sc.stateTag)
	if err != nil {
		return nil, err
	}

	if result == nil || len(result) == 0 {
		return nil, pkgerrors.New("ServiceState not found")
	}

	if result == nil {
		return nil, pkgerrors.New("Unknown Error")
	}

	state := &ServiceState{}
	err = db.DBconn.Unmarshal(result[0], state)
	if err != nil {
		return nil, err
	}

	return state, nil
}

func (sc *ServiceClient) storeServiceState(ctx context.Context, key *ServiceKey, serviceState *ServiceState) error {
	return db.DBconn.Insert(ctx, sc.storeName, key, nil, sc.stateTag, serviceState)
}

// CreateService creates a Service in the database
func (sc *ServiceClient) CreateService(ctx context.Context, service *Service) error {
	// Check if the Service already exists
	key := ServiceKey{
		Name:    service.MetaData.Name,
		Project: service.MetaData.Project,
	}

	exists, err := sc.IsServiceExists(ctx, &key)
	if err != nil {
		return err
	}

	if exists {
		return pkgerrors.New(fmt.Sprintf("Service already exists %v", key))
	}

	if service.Spec.Digs == nil {
		service.Spec.Digs = []string{}
	}

	if err = ValidateDigIds(service.Spec.Digs); err != nil {
		return err
	}

	if utils.HasDuplication(service.Spec.Digs) {
		return fmt.Errorf("invalid service spec, has duplicated DIGs")
	}

	service.MetaData.Id = uuid.New().String()
	for _, digId := range service.Spec.Digs {
		digKey, err := DeploymentIntentGroupKeyFromDigId(digId)
		if err != nil {
			return err
		}

		if digKey.Project != service.MetaData.Project {
			return fmt.Errorf("invalid DIG \"%s\" of Service \"%s\", not allowed to use DIG from a different project", digId, service.MetaData.Name)
		}

		dig, err := sc.digClient.GetDeploymentIntentGroup(ctx, digKey.Name, digKey.Project, digKey.CompositeApp, digKey.Version)
		if err != nil {
			return err
		}

		dig.addService(service.MetaData.Id)
		if err := sc.storeDig(ctx, digKey, &dig); err != nil {
			return err
		}
	}

	if err = sc.storeService(ctx, service); err != nil {
		return err
	}

	state := &ServiceState{
		ServiceActions:     []ServiceAction{{State: ActionCreated, TimeStamp: time.Now()}},
		ToInstantiatedDIGs: map[string]bool{},
	}

	return sc.storeServiceState(ctx, &key, state)
}

// UpdateService updates existing Service
func (sc *ServiceClient) UpdateService(ctx context.Context, newService *Service) error {
	key := &ServiceKey{
		Name:    newService.MetaData.Name,
		Project: newService.MetaData.Project,
	}

	currentService, err := sc.GetService(ctx, key)
	if err != nil {
		return err
	}

	if newService.Spec.Digs == nil {
		newService.Spec.Digs = []string{}
	}

	if err = ValidateDigIds(newService.Spec.Digs); err != nil {
		return err
	}

	if utils.HasDuplication(newService.Spec.Digs) {
		return fmt.Errorf("invalid service spec, has duplicated DIGs")
	}

	newService.MetaData.Id = currentService.MetaData.Id
	// Add new DIGs and remove non-existing DIGs in the new Service Spec
	newDIGs := utils.ListDifference(newService.Spec.Digs, currentService.Spec.Digs)
	removedDIGs := utils.ListDifference(currentService.Spec.Digs, newService.Spec.Digs)

	serviceState, err := sc.GetServiceState(ctx, key)
	if err != nil {
		return err
	}

	serviceState.AddToInstantiateDIGs(newDIGs)
	serviceState.RemoveToInstantiateDIGs(removedDIGs)
	err = sc.storeServiceState(ctx, key, serviceState)
	if err != nil {
		return err
	}

	currentService.MetaData.Description = newService.MetaData.Description
	if err := sc.handleRemoveDigs(ctx, currentService, removedDIGs); err != nil {
		return err
	}

	return sc.handleAddDigs(ctx, currentService, serviceState, newDIGs)
}

// GetAllServices returns all the Services
func (sc *ServiceClient) GetAllServices(ctx context.Context, project string) ([]*Service, error) {
	key := ServiceKey{
		Name:    "",
		Project: project,
	}

	//Check if project exists
	_, err := NewProjectClient().GetProject(ctx, project)
	if err != nil {
		return nil, pkgerrors.Wrap(err, "Project not found")
	}

	var serviceList []*Service
	result, err := db.DBconn.Find(ctx, sc.storeName, key, sc.tag)
	if err != nil {
		return nil, err
	}

	for _, value := range result {
		service := &Service{}
		err = db.DBconn.Unmarshal(value, service)
		if err != nil {
			return nil, err
		}

		serviceList = append(serviceList, service)
	}

	return serviceList, nil
}

// GetService returns the Service with a given service key
func (sc *ServiceClient) GetService(ctx context.Context, key *ServiceKey) (*Service, error) {
	result, err := db.DBconn.Find(ctx, sc.storeName, key, sc.tag)
	if err != nil {
		return nil, err
	}

	if result == nil || len(result) == 0 {
		return nil, pkgerrors.New("Service not found")
	}

	if result == nil {
		return nil, pkgerrors.New("Unknown Error")
	}

	service := &Service{}
	err = db.DBconn.Unmarshal(result[0], service)
	if err != nil {
		return nil, err
	}

	return service, nil
}

// IsServiceExists checks if a Service with a given service key exits
func (sc *ServiceClient) IsServiceExists(ctx context.Context, key *ServiceKey) (bool, error) {
	result, err := db.DBconn.Find(ctx, sc.storeName, key, sc.tag)
	if err != nil {
		return false, err
	}

	if result == nil || len(result) == 0 {
		return false, nil
	}

	return true, nil
}

func (sc *ServiceClient) handleRemoveDigs(ctx context.Context, service *Service, digIds []string) error {
	if err := service.verifyContainDigs(digIds); err != nil {
		return err
	}

	digsMap := map[string]bool{}
	if service.Spec.Digs == nil {
		service.Spec.Digs = []string{}
	}

	for _, dig := range service.Spec.Digs {
		digsMap[dig] = true
	}

	for _, digId := range digIds {
		delete(digsMap, digId)
		isDigInstantiated, err := sc.isServiceDigInstantiated(ctx, service, digId)
		if err != nil {
			return err
		}

		if isDigInstantiated {
			if err = sc.handleTerminateDig(ctx, service, digId); err != nil {
				return err
			}
		}

		digKey, err := DeploymentIntentGroupKeyFromDigId(digId)
		if err != nil {
			return err
		}

		dig, err := sc.digClient.GetDeploymentIntentGroup(ctx, digKey.Name, digKey.Project, digKey.CompositeApp, digKey.Version)
		if err != nil {
			return err
		}

		dig.deleteService(service.MetaData.Id)
		if err := sc.storeDig(ctx, digKey, &dig); err != nil {
			return err
		}
	}

	var newDigs []string
	for dig := range digsMap {
		newDigs = append(newDigs, dig)
	}

	service.Spec.Digs = newDigs
	return sc.storeService(ctx, service)
}

func (sc *ServiceClient) handleAddDigs(ctx context.Context, service *Service, serviceState *ServiceState, digIds []string) error {
	isInstantiatedService := serviceState.LastState().State == ActionInstantiating
	for _, digId := range digIds {
		digKey, err := DeploymentIntentGroupKeyFromDigId(digId)
		if err != nil {
			return err
		}

		if digKey.Project != service.MetaData.Project {
			return fmt.Errorf("invalid DIG \"%s\" of Service \"%s\", not allowed to use DIG from a different project", digId, service.MetaData.Name)
		}

		service.Spec.Digs = append(service.Spec.Digs, digId)
		dig, err := sc.digClient.GetDeploymentIntentGroup(ctx, digKey.Name, digKey.Project, digKey.CompositeApp, digKey.Version)
		if err != nil {
			return err
		}

		dig.addService(service.MetaData.Id)
		if err := sc.storeDig(ctx, digKey, &dig); err != nil {
			return err
		}

		if isInstantiatedService {
			if err = sc.handleInstantiateDig(ctx, service, digId); err != nil {
				return err
			}
		}
	}

	return sc.storeService(ctx, service)
}

func (sc *ServiceClient) UpdateServiceDigs(ctx context.Context, key *ServiceKey, sdu *ServiceDigsUpdate) (*Service, error) {
	err := sdu.Validate()
	if err != nil {
		return nil, err
	}

	service, err := sc.GetService(ctx, key)
	if err != nil {
		return nil, err
	}

	serviceState, err := sc.GetServiceState(ctx, key)
	if err != nil {
		return nil, err
	}

	// Add new digs
	if sdu.Add != nil {
		if utils.HasIntersection(service.Spec.Digs, sdu.Add) {
			return nil, fmt.Errorf("invalid ServiceDigUpdate, \"add\" has DIGs that is already part of service")
		}

		serviceState.AddToInstantiateDIGs(sdu.Add)
		err = sc.storeServiceState(ctx, key, serviceState)
		if err != nil {
			return nil, err
		}

		if err = sc.handleAddDigs(ctx, service, serviceState, sdu.Add); err != nil {
			return nil, err
		}
	}

	// Remove digs
	if sdu.Remove != nil {
		serviceState.RemoveToInstantiateDIGs(sdu.Remove)
		err = sc.storeServiceState(ctx, key, serviceState)
		if err != nil {
			return nil, err
		}

		if err = sc.handleRemoveDigs(ctx, service, sdu.Remove); err != nil {
			return nil, err
		}
	}

	return service, nil
}

// DeleteService deletes a Service
func (sc *ServiceClient) DeleteService(ctx context.Context, key *ServiceKey) error {
	serviceStatusInfo, err := sc.ServiceStatus(ctx, key)
	if err != nil {
		return err
	}

	if serviceStatusInfo.DeployedStatus != StatusCreated &&
		serviceStatusInfo.DeployedStatus != StatusTerminated &&
		serviceStatusInfo.DeployedStatus != StatusFailed {
		return fmt.Errorf("\"Service\" must be terminated before deleting")
	}

	service, err := sc.GetService(ctx, key)
	if err != nil {
		return err
	}

	if err = sc.handleRemoveDigs(ctx, service, service.Spec.Digs); err != nil {
		return err
	}

	return db.DBconn.Remove(ctx, sc.storeName, key)
}

// InstantiateService instantiates a Service
func (sc *ServiceClient) InstantiateService(ctx context.Context, key *ServiceKey, sda *ServiceDigsAction) error {
	service, err := sc.GetService(ctx, key)
	if err != nil {
		return err
	}

	state, err := sc.GetServiceState(ctx, key)
	if err != nil {
		return err
	}

	// Instantiate all DIGs
	//if err = service.verifyContainDigs(sda.Digs); err != nil {
	//	return err
	//}
	//
	//// If no digs provided, then apply for all DIGs
	//digs := sda.Digs
	//if digs == nil {
	//	digs = service.Spec.Digs
	//}

	digs := service.Spec.Digs
	if len(digs) == 0 {
		log.Warn("no DIGs found, skip instantiating Service %s", log.Fields{"service": service.MetaData.Name})
		return nil
	}

	if state.LastState().State == ActionInstantiating && !sda.Force {
		return fmt.Errorf("service is already started instantiating")
	}

	state.AddState(ActionInstantiating)
	state.AddToInstantiateDIGs(service.Spec.Digs)
	if err = sc.storeServiceState(ctx, key, state); err != nil {
		return err
	}

	for _, digId := range digs {
		isDigInstantiated, err := sc.isServiceDigInstantiated(ctx, service, digId)
		if err != nil {
			return err
		}

		if isDigInstantiated {
			continue
		}

		if err = sc.handleInstantiateDig(ctx, service, digId); err != nil {
			if sda.Force {
				log.Warn("skip failed to instantiate DIG in force mode", log.Fields{"service": service.MetaData.Name, "digId": digId, "error": err})
				continue
			}

			return err
		}
	}

	return nil
}

// TerminateService terminates a Service
func (sc *ServiceClient) TerminateService(ctx context.Context, key *ServiceKey, sda *ServiceDigsAction) error {
	service, err := sc.GetService(ctx, key)
	if err != nil {
		return err
	}

	state, err := sc.GetServiceState(ctx, key)
	if err != nil {
		return err
	}

	if state.LastState().State != ActionInstantiating {
		return fmt.Errorf("not allowed to terminate a non instantiated Service %s", service.MetaData.Name)
	}

	// Terminate all DIGs
	//if err = service.verifyContainDigs(sda.Digs); err != nil {
	//	return err
	//}
	//
	//// If no digs provided, then apply for all DIGs
	//digs := sda.Digs
	//if digs == nil {
	//	digs = service.Spec.Digs
	//}

	digs := service.Spec.Digs
	if len(digs) == 0 {
		log.Warn("no DIGs found, skip terminating Service", log.Fields{"service": service.MetaData.Name})
		return nil
	}

	state.ClearToInstantiateDIG()
	if err = sc.storeServiceState(ctx, key, state); err != nil {
		return err
	}

	for _, digId := range digs {
		isDigInstantiated, err := sc.isServiceDigInstantiated(ctx, service, digId)
		if err != nil {
			return err
		}

		if !isDigInstantiated {
			continue
		}

		if err = sc.handleTerminateDig(ctx, service, digId); err != nil {
			if sda.Force {
				log.Warn("skip failed to terminate DIG in force mode", log.Fields{"service": service.MetaData.Name, "digId": digId, "error": err})
				continue
			}

			return err
		}
	}

	return nil
}

// InstantiateServiceDIGs instantiates DIGs of a Service
func (sc *ServiceClient) InstantiateServiceDIGs(ctx context.Context, key *ServiceKey, sda *ServiceDigsAction) error {
	service, err := sc.GetService(ctx, key)
	if err != nil {
		return err
	}

	state, err := sc.GetServiceState(ctx, key)
	if err != nil {
		return err
	}

	if state.LastState().State != ActionInstantiating {
		return fmt.Errorf("not allowed to instantiated specific DIGs for for non instantiated Service %s", service.MetaData.Name)
	}

	if err = service.verifyContainDigs(sda.Digs); err != nil {
		return err
	}

	// If no digs provided, then apply for all DIGs
	digs := sda.Digs
	if digs == nil || len(digs) == 0 {
		return fmt.Errorf("no DIGs provided to instantiating Service %s", service.MetaData.Name)
	}

	state.AddToInstantiateDIGs(digs)
	if err = sc.storeServiceState(ctx, key, state); err != nil {
		return err
	}

	for _, digId := range digs {
		isDigInstantiated, err := sc.isServiceDigInstantiated(ctx, service, digId)
		if err != nil {
			return err
		}

		if isDigInstantiated {
			if sda.Force {
				continue
			}

			return fmt.Errorf("service %q dig %q already instantiated", service.MetaData.Name, digId)
		}

		if err = sc.handleInstantiateDig(ctx, service, digId); err != nil {
			if sda.Force {
				log.Warn("skip failed to instantiate DIG in force mode", log.Fields{"service": service.MetaData.Name, "digId": digId, "error": err})
			}

			return err
		}
	}

	return nil
}

// TerminateServiceDIGs terminates DIGs of a Service
func (sc *ServiceClient) TerminateServiceDIGs(ctx context.Context, key *ServiceKey, sda *ServiceDigsAction) error {
	service, err := sc.GetService(ctx, key)
	if err != nil {
		return err
	}

	state, err := sc.GetServiceState(ctx, key)
	if err != nil {
		return err
	}

	if err = service.verifyContainDigs(sda.Digs); err != nil {
		return err
	}

	if state.LastState().State != ActionInstantiating {
		return fmt.Errorf("not allowed to terminate DIGs for non instantiated Service %s", service.MetaData.Name)
	}

	digs := sda.Digs
	if digs == nil || len(digs) == 0 {
		return fmt.Errorf("no DIGs provided to terminate %s", service.MetaData.Name)
	}

	state.RemoveToInstantiateDIGs(digs)
	if err = sc.storeServiceState(ctx, key, state); err != nil {
		return err
	}

	for _, digId := range digs {
		isDigInstantiated, err := sc.isServiceDigInstantiated(ctx, service, digId)
		if err != nil {
			return err
		}

		if !isDigInstantiated {
			if sda.Force {
				continue
			}

			return fmt.Errorf("service %q dig %q already terminated", service.MetaData.Name, digId)
		}

		if err = sc.handleTerminateDig(ctx, service, digId); err != nil {
			if sda.Force {
				log.Warn("skip failed to terminate DIG in force mode", log.Fields{"service": service.MetaData.Name, "digId": digId, "error": err})
				continue
			}

			return err
		}
	}

	return nil
}

func (sc *ServiceClient) serviceStatus(ctx context.Context, service *Service, serviceState *ServiceState) (*ServiceStatusInfo, error) {
	status := &ServiceStatusInfo{DigsStatus: map[string]DeploymentStatus{}}

	lastAction := serviceState.LastState().State
	if lastAction != ActionCreated && lastAction != ActionInstantiating && lastAction != ActionTerminating {
		return nil, fmt.Errorf("unknown Service action %q", lastAction)
	}

	// Newly created Service
	if lastAction == ActionCreated {
		status.DeployedStatus = StatusCreated
		status.ReadyStatus = StatusNotReady

		return status, nil
	}

	instantiatedCount := 0
	isFailed := false
	isReady := true

	// Check status for all DIGs
	for digId, _ := range serviceState.ToInstantiatedDIGs {
		digKey, err := DeploymentIntentGroupKeyFromDigId(digId)
		if err != nil {
			return nil, err
		}

		dig, err := sc.digClient.GetDeploymentIntentGroup(ctx, digKey.Name, digKey.Project, digKey.CompositeApp, digKey.Version)
		if err != nil {
			return nil, err
		}

		digStatus, err := NewInstantiationClient().Status(ctx, digKey.Project, digKey.CompositeApp, digKey.Version,
			digKey.Name, "", "ready", "all", nil, nil, nil)
		if err != nil {
			return nil, err
		}

		status.DigsStatus[digId] = digStatus
		_, digServiceInstantiated := dig.Spec.InstantiatedServices[service.MetaData.Id]

		if digStatus.DeployedStatus == appcontext.AppContextStatusEnum.InstantiateFailed ||
			digStatus.DeployedStatus == appcontext.AppContextStatusEnum.TerminateFailed ||
			digStatus.DeployedStatus == appcontext.AppContextStatusEnum.UpdateFailed {
			// Failed DIG
			isFailed = true
			isReady = false
		} else if digStatus.DeployedStatus == appcontext.AppContextStatusEnum.Instantiated && digServiceInstantiated {
			// Instantiated DIG
			instantiatedCount += 1
			isReady = isReady && digStatus.ReadyStatus == StatusReady
		}
	}

	status.ReadyStatus = StatusNotReady

	if isFailed {
		status.DeployedStatus = StatusFailed
	} else if lastAction == ActionInstantiating && instantiatedCount == len(serviceState.ToInstantiatedDIGs) {
		status.DeployedStatus = StatusInstantiated
		if isReady {
			status.ReadyStatus = StatusReady
		}
	} else if lastAction == ActionTerminating && instantiatedCount == 0 {
		status.DeployedStatus = StatusTerminated
	} else {
		status.DeployedStatus = StatusProcessing
	}

	return status, nil
}

func (sc *ServiceClient) ServiceStatus(ctx context.Context, key *ServiceKey) (*ServiceStatusInfo, error) {
	service, err := sc.GetService(ctx, key)
	if err != nil {
		return nil, err
	}

	serviceState, err := sc.GetServiceState(ctx, key)
	if err != nil {
		return nil, err
	}

	return sc.serviceStatus(ctx, service, serviceState)
}

func (sc *ServiceClient) instantiateDig(ctx context.Context, digKey *DeploymentIntentGroupKey) error {
	if err := sc.InstantiationClient.Approve(ctx, digKey.Project, digKey.CompositeApp, digKey.Version, digKey.Name); err != nil {
		return err
	}

	return sc.InstantiationClient.Instantiate(ctx, digKey.Project, digKey.CompositeApp, digKey.Version, digKey.Name)
}

func (sc *ServiceClient) terminateDig(ctx context.Context, digKey *DeploymentIntentGroupKey) error {
	return sc.InstantiationClient.Terminate(ctx, digKey.Project, digKey.CompositeApp, digKey.Version, digKey.Name)
}

func (sc *ServiceClient) handleInstantiateDig(ctx context.Context, service *Service, digId string) error {
	digKey, err := DeploymentIntentGroupKeyFromDigId(digId)
	if err != nil {
		return err
	}

	dig, err := sc.digClient.GetDeploymentIntentGroup(ctx, digKey.Name, digKey.Project, digKey.CompositeApp, digKey.Version)
	if err != nil {
		return err
	}

	if dig.Spec.InstantiatedServices == nil {
		dig.Spec.InstantiatedServices = map[string]interface{}{}
	}

	digState, err := sc.digClient.GetDeploymentIntentGroupState(ctx, digKey.Name, digKey.Project, digKey.CompositeApp, digKey.Version)
	if err != nil {
		return err
	}

	digLastAction := digState.Actions[len(digState.Actions)-1].State

	isDigInstantiated := len(dig.Spec.InstantiatedServices) > 0 || digLastAction == state.StateEnum.Instantiated
	dig.Spec.InstantiatedServices[service.MetaData.Id] = true
	if err = sc.storeDig(ctx, digKey, &dig); err != nil {
		return err
	}
	if isDigInstantiated {
		return sc.InstantiationClient.UpdateInstantiated(ctx, digKey.Project, digKey.CompositeApp, digKey.Version, digKey.Name)
	}

	return sc.instantiateDig(ctx, digKey)
}

func (sc *ServiceClient) handleTerminateDig(ctx context.Context, service *Service, digId string) error {
	digKey, err := DeploymentIntentGroupKeyFromDigId(digId)
	if err != nil {
		return err
	}

	dig, err := sc.digClient.GetDeploymentIntentGroup(ctx, digKey.Name, digKey.Project, digKey.CompositeApp, digKey.Version)
	if err != nil {
		return err
	}

	if dig.Spec.InstantiatedServices == nil || len(dig.Spec.InstantiatedServices) == 0 {
		return fmt.Errorf("terminating uninstantiated dig %q for service %q", digId, service.MetaData.Name)
	}

	isOnlyService := len(dig.Spec.InstantiatedServices) == 1
	delete(dig.Spec.InstantiatedServices, service.MetaData.Id)
	if err = sc.storeDig(ctx, digKey, &dig); err != nil {
		return err
	}

	if !isOnlyService {
		return sc.InstantiationClient.UpdateInstantiated(ctx, digKey.Project, digKey.CompositeApp, digKey.Version, digKey.Name)
	}

	return sc.terminateDig(ctx, digKey)
}

func (sc *ServiceClient) isServiceDigInstantiated(ctx context.Context, service *Service, digId string) (bool, error) {
	digKey, err := DeploymentIntentGroupKeyFromDigId(digId)
	if err != nil {
		return false, err
	}

	dig, err := sc.digClient.GetDeploymentIntentGroup(ctx, digKey.Name, digKey.Project, digKey.CompositeApp, digKey.Version)
	if err != nil {
		return false, err
	}

	if dig.Spec.InstantiatedServices == nil {
		return false, nil
	}

	_, ok := dig.Spec.InstantiatedServices[service.MetaData.Id]
	return ok, nil
}
