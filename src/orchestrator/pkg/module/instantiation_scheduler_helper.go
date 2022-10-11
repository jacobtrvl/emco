// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package module

import (
	"container/heap"
	"context"
	"strings"

	"fmt"

	pkgerrors "github.com/pkg/errors"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/appcontext"
	client "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/grpc/contextupdateclient"
	rsyncclient "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/grpc/installappclient"
	plsGrpcClient "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/grpc/placementcontrollerclient"
	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/module/controller"
	mtypes "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/module/types"
	"go.opentelemetry.io/otel/trace"
)

// ControllerTypePlacement denotes "placement" Controller Type
const ControllerTypePlacement string = "placement"

// ControllerTypeAction denotes "action" Controller Type
const ControllerTypeAction string = "action"

// rsyncName denotes the name of the rsync controller
const rsyncName = "rsync"

// ControllerElement consists of controller and an internal field - index
type ControllerElement struct {
	controller controller.Controller
	index      int // used for indexing the HeapArray
}

// PrioritizedControlList contains PrioritizedList of PlacementControllers and ActionControllers
type PrioritizedControlList struct {
	pPlaCont []controller.Controller
	pActCont []controller.Controller
}

// PriorityQueue is the heapArray to store the Controllers
type PriorityQueue []*ControllerElement

func (pq PriorityQueue) Len() int { return len(pq) }

func (pq PriorityQueue) Less(i, j int) bool {
	// We want Pop to give us highest Priority controller
	// The lower the number, higher the priority
	return pq[i].controller.Spec.Priority < pq[j].controller.Spec.Priority
}

// Pop method returns the controller with the highest priority
func (pq *PriorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	c := old[n-1]
	c.index = -1
	*pq = old[0 : n-1]
	return c
}

// Push method add a controller into the heapArray
func (pq *PriorityQueue) Push(c interface{}) {
	n := len(*pq)
	controllerElement := c.(*ControllerElement)
	controllerElement.index = n
	*pq = append(*pq, controllerElement)
}

func (pq PriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}

func getPrioritizedControllerList(ctx context.Context, p, ca, v, di string) (PrioritizedControlList, map[string]string, error) {
	listOfControllers := make([]string, 0) // shall contain the real controllerNames to be passed to controllerAPI
	mapOfControllers := make(map[string]string)

	iList, err := NewIntentClient().GetAllIntents(ctx, p, ca, v, di)
	if err != nil {
		return PrioritizedControlList{}, map[string]string{}, err
	}
	for _, eachmap := range iList.ListOfIntents {
		for controller, controllerIntent := range eachmap {
			if controller != GenericPlacementIntentName {
				listOfControllers = append(listOfControllers, controller)
				mapOfControllers[controller] = controllerIntent
			}
		}
	}

	listPC := make([]*ControllerElement, 0)
	listAC := make([]*ControllerElement, 0)

	log.Info("getPrioritizedControllerList .. controllers info", log.Fields{"listOfControllers": listOfControllers})
	for _, cn := range listOfControllers {
		c, err := NewClient().Controller.GetController(ctx, cn)

		if err != nil {
			return PrioritizedControlList{}, map[string]string{}, err
		}
		log.Info("getPrioritizedControllerList .. controllers info", log.Fields{"Spec.Type": c.Spec.Type})
		if c.Spec.Type == ControllerTypePlacement {
			// Collect in listPC
			listPC = append(listPC, &ControllerElement{controller: controller.Controller{
				Metadata: mtypes.Metadata{
					Name:        c.Metadata.Name,
					Description: c.Metadata.Description,
					UserData1:   c.Metadata.UserData1,
					UserData2:   c.Metadata.UserData2,
				},
				Spec: controller.ControllerSpec{
					Host:     c.Spec.Host,
					Port:     c.Spec.Port,
					Type:     c.Spec.Type,
					Priority: c.Spec.Priority,
				},
			}})
		} else if c.Spec.Type == ControllerTypeAction {
			// Collect in listAC
			listAC = append(listAC, &ControllerElement{controller: controller.Controller{
				Metadata: mtypes.Metadata{
					Name:        c.Metadata.Name,
					Description: c.Metadata.Description,
					UserData1:   c.Metadata.UserData1,
					UserData2:   c.Metadata.UserData2,
				},
				Spec: controller.ControllerSpec{
					Host:     c.Spec.Host,
					Port:     c.Spec.Port,
					Type:     c.Spec.Type,
					Priority: c.Spec.Priority,
				},
			}})
		} else {
			log.Info("Controller type undefined", log.Fields{"Controller type": c.Spec.Type, "ControllerName": c.Metadata.Name})
		}
	}

	pqPlacementCont := make(PriorityQueue, len(listPC))
	for i, eachPC := range listPC {
		pqPlacementCont[i] = &ControllerElement{controller: eachPC.controller, index: i}
	}
	prioritizedPlaControllerList := make([]controller.Controller, 0)
	heap.Init(&pqPlacementCont)
	for pqPlacementCont.Len() > 0 {
		ce := heap.Pop(&pqPlacementCont).(*ControllerElement)

		prioritizedPlaControllerList = append(prioritizedPlaControllerList, ce.controller)
	}

	pqActionCont := make(PriorityQueue, len(listAC))
	for i, eachAC := range listAC {
		pqActionCont[i] = &ControllerElement{controller: eachAC.controller, index: i}
	}
	prioritizedActControllerList := make([]controller.Controller, 0)
	heap.Init(&pqActionCont)
	for pqActionCont.Len() > 0 {
		ce := heap.Pop(&pqActionCont).(*ControllerElement)
		prioritizedActControllerList = append(prioritizedActControllerList, ce.controller)
	}

	log.Info("getPrioritizedControllerList .. controllers info", log.Fields{"placement-controllers": prioritizedPlaControllerList, "action-controllers": prioritizedActControllerList})
	prioritizedControlList := PrioritizedControlList{pPlaCont: prioritizedPlaControllerList, pActCont: prioritizedActControllerList}

	return prioritizedControlList, mapOfControllers, nil

}

/*
callGrpcForControllerList method shall take in a list of controllers, a map of contollers to controllerIntentNames and contextID. It invokes the context
updation through the grpc client for the given list of controllers.
*/
func callGrpcForControllerList(ctx context.Context, cl []controller.Controller, mc map[string]string, contextid, updateFromContextid interface{}) error {
	for _, c := range cl {
		controller := c.Metadata.Name
		controllerIntentName := mc[controller]
		appContextID := fmt.Sprintf("%v", contextid)
		updateAppContextId := fmt.Sprintf("%v", updateFromContextid)
		log.Info("callGrpcForControllerList .. Invoking action-controller.", log.Fields{
			"controller": controller, "controllerIntentName": controllerIntentName, "appContextID": appContextID})
		err := client.InvokeContextUpdate(ctx, controller, controllerIntentName, appContextID, updateAppContextId)
		if err != nil {
			return err
		}
	}
	return nil
}

/*
callGrpcForPlacementControllerList method shall take in a list of placement controllers, a map of contollers to controllerIntentNames and contextID.
It invokes the filter clusters through the grpc client for the given list of controllers.
*/
func callGrpcForPlacementControllerList(ctx context.Context, cl []controller.Controller, contextid interface{}) error {
	for _, c := range cl {
		controller := c.Metadata.Name
		appContextID := fmt.Sprintf("%v", contextid)
		log.Info("callGrpcForControllerList .. Invoking placement-controller.", log.Fields{
			"controller": controller, "appContextID": appContextID})
		err := plsGrpcClient.InvokeFilterClusters(ctx, c, appContextID)
		if err != nil {
			return pkgerrors.Wrapf(err, "Placement-controller returned error. failed-placement-controller[%v] appContextID[%v]", controller, appContextID)
		}
	}
	return nil
}

/*
queryDBAndSetRsyncInfo queries the MCO db to find the record the sync controller
and then sets the RsyncInfo global variable.
*/
func queryDBAndSetRsyncInfo(ctx context.Context) (rsyncclient.RsyncInfo, error) {
	client := controller.NewControllerClient("resources", "data", "orchestrator")
	vals, _ := client.GetControllers(ctx)
	for _, v := range vals {
		if v.Metadata.Name == rsyncName {
			log.Info("Initializing RPC connection to resource synchronizer", log.Fields{
				"Controller": v.Metadata.Name,
			})
			rsyncInfo := rsyncclient.NewRsyncInfo(v.Metadata.Name, v.Spec.Host, v.Spec.Port)
			return rsyncInfo, nil
		}
	}
	return rsyncclient.RsyncInfo{}, pkgerrors.Errorf("queryRsyncInfoInMCODB Failed - Could not get find rsync by name : %v", rsyncName)
}

/*
callRsyncInstall method shall take in the app context id and invokes the rsync service via grpc
*/
func callRsyncInstall(ctx context.Context, contextid interface{}) error {
	span := trace.SpanFromContext(ctx)
	span.AddEvent("invoke-rsync")

	rsyncInfo, err := queryDBAndSetRsyncInfo(ctx)
	log.Info("Calling the Rsync ", log.Fields{
		"RsyncName": rsyncInfo.RsyncName,
	})
	if err != nil {
		return err
	}

	appContextID := fmt.Sprintf("%v", contextid)
	err = rsyncclient.InvokeInstallApp(ctx, appContextID)
	if err != nil {
		return err
	}
	return nil
}

/*
callRsyncUninstall method shall take in the app context id and invokes the rsync service via grpc
*/
func callRsyncUninstall(ctx context.Context, contextid interface{}) error {
	rsyncInfo, err := queryDBAndSetRsyncInfo(ctx)
	log.Info("Calling the Rsync ", log.Fields{
		"RsyncName": rsyncInfo.RsyncName,
	})
	if err != nil {
		return err
	}

	appContextID := fmt.Sprintf("%v", contextid)
	err = rsyncclient.InvokeUninstallApp(ctx, appContextID)
	if err != nil {
		return err
	}
	return nil
}

/*
deleteExtraClusters method shall delete the extra cluster handles for each AnyOf cluster present in the etcd after the grpc call for context updation.
*/
func deleteExtraClusters(ctx context.Context, apps []App, ct appcontext.AppContext) error {
	for _, app := range apps {
		an := app.Metadata.Name
		log.Warn("", log.Fields{"an": an})
		gmap, err := ct.GetClusterGroupMap(ctx, an)
		log.Warn("", log.Fields{"gmap": gmap})
		if err != nil {
			return err
		}
		for gr, cl := range gmap {
			log.Warn("", log.Fields{"cl": cl})
			for i, cn := range cl {
				log.Warn("", log.Fields{"i": i})
				log.Warn("", log.Fields{"cn": cn})
				// avoids deleting the first cluster
				if i > 0 {
					ch, err := ct.GetClusterHandle(ctx, an, cn)
					log.Warn("", log.Fields{"err": err})
					if err != nil {
						return err
					}
					err = ct.DeleteCluster(ctx, ch)
					if err != nil {
						return err
					}
					log.Info("::Deleted cluster for::", log.Fields{"appName": an, "GroupNumber": gr, "ClusterName": cn})
				}
			}

		}
	}
	return nil
}

// callScheduler instantiates based on the controller priority list
func callScheduler(ctx context.Context, appCtx appcontext.AppContext, ctxval, ctxUpdateFromval interface{}, p, ca, v, di string) error {
	span := trace.SpanFromContext(ctx)

	// BEGIN: scheduler code
	span.AddEvent("retrieve-controllers")
	allApps, err := NewAppClient().GetApps(ctx, p, ca, v)
	if err != nil {
		return pkgerrors.Wrap(err, "Error in getting all Apps")
	}

	pl, mapOfControllers, err := getPrioritizedControllerList(ctx, p, ca, v, di)
	if err != nil {
		return pkgerrors.Wrap(err, "Error adding getting prioritized controller list")
	}

	log.Info("Orchestrator Instantiate .. Priority Based List ", log.Fields{"PlacementControllers::": pl.pPlaCont,
		"ActionControllers::": pl.pActCont, "mapOfControllers::": mapOfControllers})

	// Invoke all Placement Controllers communication interface in loop
	span.AddEvent("invoke-placement-controllers")
	err = callGrpcForPlacementControllerList(ctx, pl.pPlaCont, ctxval)
	if err != nil {
		deleteAppContext(ctx, appCtx)
		log.Error("Orchestrator Instantiate .. Error calling PlacementController gRPC.", log.Fields{"all-placement-controllers": pl.pPlaCont, "err": err})
		return pkgerrors.Wrap(err, "Error calling PlacementController gRPC")
	}

	// delete extra clusters from group map
	err = deleteExtraClusters(ctx, allApps, appCtx)
	if err != nil {
		deleteAppContext(ctx, appCtx)
		return pkgerrors.Wrap(err, "Error deleting extra clusters")
	}

	// Invoke all Action Controllers communication interface
	span.AddEvent("invoke-action-controllers")
	err = callGrpcForControllerList(ctx, pl.pActCont, mapOfControllers, ctxval, ctxUpdateFromval)
	log.Warn("", log.Fields{"pl.pActCont::": pl.pActCont})
	log.Warn("", log.Fields{"mapOfControllers::": mapOfControllers})
	log.Warn("", log.Fields{"ctxval::": ctxval})
	if err != nil {
		deleteAppContext(ctx, appCtx)
		return pkgerrors.Wrap(err, "Error calling gRPC for action controller list")
	}
	// END: Scheduler code
	return nil
}

// callScheduler terminates based on the controller priority list
func callTerminateScheduler(ctx context.Context, ctxval interface{}, p, ca, v, di string) error {

	pl, mc, err := getPrioritizedControllerList(ctx, p, ca, v, di)
	if err != nil {
		return pkgerrors.Wrap(err, "Error adding getting prioritized controller list")
	}

	for _, c := range pl.pActCont {
		controller := c.Metadata.Name
		controllerIntentName := mc[controller]
		appContextID := fmt.Sprintf("%v", ctxval)
		log.Info("callTerminateScheduler .. Invoking action-controller.", log.Fields{
			"controller": controller, "controllerIntentName": controllerIntentName, "appContextID": appContextID})
		err := client.InvokeContextTerminate(ctx, controller, appContextID)
		// If GRPC endpoint not implemented by controller don't consider that as an error
		if err != nil && !strings.Contains(err.Error(), "TerminateAppContext not implemented") {
			log.Error("InvokeContextTerminate: Error", log.Fields{"controller": controller, "err": err})
			return err
		}

	}
	return nil

}

// callPostEventScheduler instantiates based on the controller priority list
func callPostEventScheduler(ctx context.Context, ctxval interface{}, p, ca, v, di, event string) error {

	iList, err := NewIntentClient().GetAllIntents(ctx, p, ca, v, di)
	if err != nil {
		return err
	}
	appContextID := fmt.Sprintf("%v", ctxval)
	for _, eachmap := range iList.ListOfIntents {
		for controller, _ := range eachmap {
			if controller != GenericPlacementIntentName {
				log.Info("Invoking action-controller for Post Event hook.", log.Fields{
					"controller": controller, "appContextID": appContextID, "event": event})
				err := client.InvokePostEvent(ctx, controller, appContextID, event)
				// If GRPC endpoint not implemented by controller don't consider that as an error
				if err != nil && !strings.Contains(err.Error(), "PostEvent not implemented") {
					log.Error("InvokePostEvent: Error", log.Fields{"controller": controller, "err": err})
					return err
				}
			}
		}
	}
	return nil
}
