package event

import (
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/db"
)

type Client struct {
	db db.Store
}

type Event struct {
	Agent string `json:"agent"`
	Id    string `json:"id"`
}

func NewClient(db db.Store) *Client {
	return &Client{db: db}
}

type Request struct {
	Dummy string
}
