package intent

import (
	"context"
)

func NewClient(config Config) *Client {
	return &Client{
		db:           config.Db,
		tag:          config.Tag,
		storeName:    config.StoreName,
		updateStream: config.UpdateStream,
	}
}

func (c Client) CreateIntent(_ context.Context, request *Request) (*Intent, error) {
	key := Key{
		PolicyIntent:        request.PolicyIntentId,
		Project:             request.Project,
		CompositeApp:        request.CompositeApp,
		CompositeAppVersion: request.CompositeAppVersion,
		DigName:             request.DeploymentIntentGroup,
	}
	intent := *request.IntentData
	if err := c.db.Insert(c.storeName, key, nil, c.tag, intent); err != nil {
		return nil, err
	}
	return &intent, nil
}

func (c Client) DeleteIntent(_ context.Context, request *Request) error {
	key := Key{
		PolicyIntent:        request.PolicyIntentId,
		Project:             request.Project,
		CompositeApp:        request.CompositeApp,
		CompositeAppVersion: request.CompositeAppVersion,
		DigName:             request.DeploymentIntentGroup,
	}
	return c.db.Remove(c.storeName, key)
}

func (c Client) GetIntent(_ context.Context, request *Request) (*Intent, error) {
	key := Key{
		PolicyIntent:        request.PolicyIntentId,
		Project:             request.Project,
		CompositeApp:        request.CompositeApp,
		CompositeAppVersion: request.CompositeAppVersion,
		DigName:             request.DeploymentIntentGroup,
	}

	value, err := c.db.Find(c.storeName, key, c.tag)
	if err != nil || len(value) == 0 {
		return nil, err
	}
	data := new(Intent)
	if err := c.db.Unmarshal(value[0], data); err != nil {
		return nil, err
	}
	return data, nil
}