package event

import (
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/db"
)

type Client struct {
	db          db.Store
	tag         string
	storeName   string
	eventStream chan Event
}

type Event struct {
	Id      string `json:"id"`
	AgentID string `json:"agent,omitempty"`
}

func NewClient(config Config) *Client {
	return &Client{
		db:          config.Db,
		tag:         config.Tag,
		storeName:   config.StoreName,
		eventStream: config.EventStream,
	}
}

type Request struct {
	Dummy string
}

type Config struct {
	Db          db.Store
	Tag         string
	StoreName   string
	EventStream chan Event
}
