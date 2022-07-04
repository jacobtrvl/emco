package event

import (
	"emcopolicy/pkg/plugins"
	"github.com/pkg/errors"
	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
)

func DoAction(actorName string, input []byte) error {
	var (
		actor Actor
	)
	switch actorName {
	case "temporal":
		actor = new(plugins.TemporalActor)
	default:
		log.Error("DoAction: No Actor Plugin matched", log.Fields{"actor": actorName})
		return errors.Errorf("DoAction: No Actor Plugin matched with provided actor name: %s", actorName)
	}
	return actor.Execute(input)
}
