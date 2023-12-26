package taskbiz

import (
	"bestHabit/common"
	"bestHabit/modules/task/taskmodel"
	"bestHabit/pubsub"
	"context"
)

type CreateTaskStore interface {
	Create(ctx context.Context, data *taskmodel.TaskCreate) error
}

type createTaskBiz struct {
	store  CreateTaskStore
	pubsub pubsub.Pubsub
}

func NewCreateTaskBiz(store CreateTaskStore, pubsub pubsub.Pubsub) *createTaskBiz {
	return &createTaskBiz{store: store, pubsub: pubsub}
}

func (b *createTaskBiz) CreateTask(ctx context.Context, data *taskmodel.TaskCreate, userId int) error {
	if err := data.Validate(); err != nil {
		return err
	}

	data.UserId = userId

	if err := b.store.Create(ctx, data); err != nil {
		return err
	}
	go func() {
		defer common.AppRecover()
		b.pubsub.Publish(ctx, common.TopicUserCreateNewTask, pubsub.NewMessage(data))
	}()
	return nil
}
