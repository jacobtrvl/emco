package plugins

import (
	"fmt"
)

type TemporalActor struct {
}

func (t *TemporalActor) Execute(input []byte) error {
	fmt.Println("In Temporal actor in plugin", string(input))
	return nil
}
