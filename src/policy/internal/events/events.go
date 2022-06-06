package event

import (
	"context"
)

func (c Client) CreateEvent(_ context.Context, request *Request) string {
	return request.Dummy
}

func (c Client) UpdateEvent(_ context.Context, request *Request) string {
	return request.Dummy
}

func (c Client) DeleteEvent(_ context.Context, request *Request) string {
	return request.Dummy
}

func (c Client) GetEvent(_ context.Context, request *Request) string {
	return request.Dummy
}
