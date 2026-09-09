package event

import "context"

type Publisher interface {
	PublishTaskCreated(
		ctx context.Context,
		event TaskCreatedEvent,
	) error
}
