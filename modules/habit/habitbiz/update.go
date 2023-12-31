package habitbiz

import (
	"bestHabit/common"
	"bestHabit/modules/habit/habitmodel"
	"bestHabit/modules/task/taskmodel"
	"bestHabit/pubsub"
	"context"
)

type UpdateHabitStorage interface {
	FindHabitById(ctx context.Context, id int) (*habitmodel.HabitFind, error)
	UpdateHabitInfo(ctx context.Context, newInfo *habitmodel.HabitUpdate, id int) error
}

type updateHabitBiz struct {
	store  UpdateHabitStorage
	pubsub pubsub.Pubsub
}

func NewUpdateHabitBiz(store UpdateHabitStorage, pubsub pubsub.Pubsub) *updateHabitBiz {
	return &updateHabitBiz{store: store, pubsub: pubsub}
}

func (b *updateHabitBiz) Update(ctx context.Context, newInfo *habitmodel.HabitUpdate, id int) error {
	oldData, err := b.store.FindHabitById(ctx, id)

	isNeedUpdateCronJob := true

	if err != nil {
		if err == common.ErrorNoRows {
			return common.ErrEntityNotFound(habitmodel.EntityName, err)
		}

		return err
	}

	if oldData.Status == 0 {
		return common.ErrEntityDeleted(habitmodel.EntityName, err)
	}

	if newInfo.Name == nil {
		newInfo.Name = &oldData.Name
	}

	if newInfo.Description == nil {
		newInfo.Description = &oldData.Description
	}

	if newInfo.StartDate == nil && newInfo.Reminder == nil && newInfo.EndDate == nil {
		isNeedUpdateCronJob = false
	}

	if newInfo.StartDate == nil {
		newInfo.StartDate = &oldData.StartDate
	}

	if newInfo.EndDate == nil {
		newInfo.EndDate = &oldData.EndDate
	}

	if newInfo.Reminder == nil {
		newInfo.Reminder = &oldData.Reminder
	}

	if newInfo.IsCountBased == nil {
		newInfo.IsCountBased = &oldData.IsCountBased
	}

	if newInfo.Type == nil {
		newInfo.Type = &oldData.Type
	}

	if newInfo.Days == nil {
		newInfo.Days = oldData.Days
	}

	if newInfo.Target == nil {
		newInfo.Target = oldData.Target
	}

	newInfo.UserId = &oldData.UserId
	newInfo.Id = &id

	err = b.store.UpdateHabitInfo(ctx, newInfo, id)

	if err != nil {
		return common.ErrCannotUpdateEntity(taskmodel.EntityName, err)
	}

	if isNeedUpdateCronJob {
		go func() {
			defer common.AppRecover()
			b.pubsub.Publish(ctx, common.TopicUserUpdateHabit, pubsub.NewMessage(newInfo))
		}()
	}

	return nil
}
