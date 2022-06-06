package policy

import (
	"context"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/db"
)

func NewClient(db db.Store) *Client {
	return &Client{
		db:        db,
		tag:       "data",
		storeName: "resources",
	}
}

func (c Client) CreatePolicy(_ context.Context, request *PolicyRequest) (*Policy, error) {
	key := PolicyKey{
		request.PolicyId,
	}
	policy := Policy{
		Metadata: Metadata{},
		Spec:     request.Policy.Spec,
	}
	if err := c.db.Insert(c.storeName, key, nil, c.tag, policy); err != nil {
		return nil, err
	}
	return &policy, nil
}

func (c Client) DeletePolicy(_ context.Context, request *PolicyRequest) error {
	key := PolicyKey{
		request.PolicyId,
	}
	return c.db.Remove(c.storeName, key)
}

func (c Client) GetPolicy(_ context.Context, request *PolicyRequest) (*Policy, error) {
	key := PolicyKey{
		request.PolicyId,
	}
	value, err := c.db.Find(c.storeName, key, c.tag)
	if err != nil {
		return nil, err
	}
	data := new(Policy)
	if err := c.db.Unmarshal(value[0], data); err != nil {
		return nil, err
	}
	return data, err
}
