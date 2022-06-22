package context

import (
	"context"
	"fmt"
	"time"

	"gitlab.com/project-emco/core/emco-base/src/rsync/pkg/internal/utils"
	. "gitlab.com/project-emco/core/emco-base/src/rsync/pkg/types"
)

type DependentCtx struct {
	acID  string
	acRef utils.AppContextReference
}

func NewDependentCtx(cid string) *DependentCtx {
	var err error

	dc := DependentCtx{
		acID: cid,
	}

	dc.acRef, err = utils.NewAppContextReference(cid)
	if err != nil {
		return nil
	}
	return &dc
}

func (dc *DependentCtx) PropagateStop() error {
	return nil
}

func (dc *DependentCtx) WaitToComplete(ctx context.Context, op RsyncOperation, app, cluster string) {
	l, err := dc.acRef.GetAllDependentCtx()

	if err != nil {
		return
	}
	for cid, operation := range l {
		if operation != fmt.Sprint(op) {
			continue
		}
		ref, err := utils.NewAppContextReference(cid)
		if err != nil {
			// On error ignore dependency
			continue
		}

		match, err := dc.MatchAppClusterState(op, ref, app, cluster)
		if err != nil {
			// On error ignore dependency
			continue
		}
		if !match {
		Loop:
			for {
				select {
				// Wait before checking if resource is ready
				case <-time.After(time.Duration(10) * time.Millisecond):
					// Context is canceled
					if ctx.Err() != nil {
						return
					}
					match, err := dc.MatchAppClusterState(op, ref, app, cluster)
					if match || err != nil {
						break Loop
					} else {
						continue
					}
				case <-ctx.Done():
					return
				}
			}
		}
	}
	return
}

// If the cluster state matches RsyncOperation return true
func (dc *DependentCtx) MatchAppClusterState(op RsyncOperation, ref utils.AppContextReference, app, cluster string) (bool, error) {

	state, err := ref.GetClusterResourcesState(app, cluster)
	if err != nil {
		return false, err
	}
	// No state
	if state == "" {
		return false, nil
	}

	if state != fmt.Sprint(op) {
		return false, nil
	}
	return true, nil
}
